import Ember from 'ember';
import { task } from 'ember-concurrency';
import lazyCapabilities, { apiPath } from 'vault/macros/lazy-capabilities';

const { Service, computed, inject } = Ember;
const ROOT_NAMESPACE = '';
export default Service.extend({
  store: inject.service(),
  auth: inject.service(),
  userRootNamespace: computed.alias('auth.authData.userRootNamespace'),
  //populated by the query param on the cluster route
  path: null,
  // list of namespaces available to the current user under the
  // current namespace
  accessibleNamespaces: null,

  inRootNamespace: computed.equal('path', ROOT_NAMESPACE),

  setNamespace(path) {
    this.set('path', path);
  },
  listPath: lazyCapabilities(apiPath`sys/namespaces/`, 'path'),
  canList: computed.alias('listPath.canList'),

  findNamespacesForUser: task(function*() {
    // uses the adapter and the raw response here since
    // models get wiped when switching namespaces and we
    // want to keep track of these separately
    let store = this.get('store');
    let adapter = store.adapterFor('namespace');
    let userRoot = this.get('auth.authData.userRootNamespace');
    try {
      let ns = yield adapter.findAll(store, 'namespace', null, {
        adapterOptions: {
          forUser: true,
          namespace: userRoot,
        },
      });
      this.set(
        'accessibleNamespaces',
        ns.data.keys.map(n => {
          let fullNS = n;
          // if the user's root isn't '', then we need to construct
          // the paths so they connect to the user root to the list
          // otherwise using the current ns to grab the correct leaf
          // node in the graph doesn't work
          if (userRoot) {
            fullNS = `${userRoot}/${n}`;
          }
          return fullNS.replace(/\/$/, '');
        })
      );
    } catch (e) {
      //do nothing here
    }
  }).drop(),
});
