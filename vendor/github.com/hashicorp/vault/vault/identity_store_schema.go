package vault

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
)

const (
	entitiesTable      = "entities"
	entityAliasesTable = "entity_aliases"
	groupsTable        = "groups"
	groupAliasesTable  = "group_aliases"
)

func identityStoreSchema() *memdb.DBSchema {
	iStoreSchema := &memdb.DBSchema{
		Tables: make(map[string]*memdb.TableSchema),
	}

	schemas := []func() *memdb.TableSchema{
		entitiesTableSchema,
		aliasesTableSchema,
		groupsTableSchema,
		groupAliasesTableSchema,
	}

	for _, schemaFunc := range schemas {
		schema := schemaFunc()
		if _, ok := iStoreSchema.Tables[schema.Name]; ok {
			panic(fmt.Sprintf("duplicate table name: %s", schema.Name))
		}
		iStoreSchema.Tables[schema.Name] = schema
	}

	return iStoreSchema
}

func aliasesTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: entityAliasesTable,
		Indexes: map[string]*memdb.IndexSchema{
			"id": &memdb.IndexSchema{
				Name:   "id",
				Unique: true,
				Indexer: &memdb.StringFieldIndex{
					Field: "ID",
				},
			},
			"factors": &memdb.IndexSchema{
				Name:   "factors",
				Unique: true,
				Indexer: &memdb.CompoundIndex{
					Indexes: []memdb.Indexer{
						&memdb.StringFieldIndex{
							Field: "MountAccessor",
						},
						&memdb.StringFieldIndex{
							Field: "Name",
						},
					},
				},
			},
		},
	}
}

func entitiesTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: entitiesTable,
		Indexes: map[string]*memdb.IndexSchema{
			"id": &memdb.IndexSchema{
				Name:   "id",
				Unique: true,
				Indexer: &memdb.StringFieldIndex{
					Field: "ID",
				},
			},
			"name": &memdb.IndexSchema{
				Name:   "name",
				Unique: true,
				Indexer: &memdb.StringFieldIndex{
					Field: "Name",
				},
			},
			"merged_entity_ids": &memdb.IndexSchema{
				Name:         "merged_entity_ids",
				Unique:       true,
				AllowMissing: true,
				Indexer: &memdb.StringSliceFieldIndex{
					Field: "MergedEntityIDs",
				},
			},
			"bucket_key_hash": &memdb.IndexSchema{
				Name:         "bucket_key_hash",
				Unique:       false,
				AllowMissing: false,
				Indexer: &memdb.StringFieldIndex{
					Field: "BucketKeyHash",
				},
			},
		},
	}
}

func groupsTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: groupsTable,
		Indexes: map[string]*memdb.IndexSchema{
			"id": {
				Name:   "id",
				Unique: true,
				Indexer: &memdb.StringFieldIndex{
					Field: "ID",
				},
			},
			"name": {
				Name:   "name",
				Unique: true,
				Indexer: &memdb.StringFieldIndex{
					Field: "Name",
				},
			},
			"member_entity_ids": {
				Name:         "member_entity_ids",
				Unique:       false,
				AllowMissing: true,
				Indexer: &memdb.StringSliceFieldIndex{
					Field: "MemberEntityIDs",
				},
			},
			"parent_group_ids": {
				Name:         "parent_group_ids",
				Unique:       false,
				AllowMissing: true,
				Indexer: &memdb.StringSliceFieldIndex{
					Field: "ParentGroupIDs",
				},
			},
			"bucket_key_hash": &memdb.IndexSchema{
				Name:         "bucket_key_hash",
				Unique:       false,
				AllowMissing: false,
				Indexer: &memdb.StringFieldIndex{
					Field: "BucketKeyHash",
				},
			},
		},
	}
}

func groupAliasesTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: groupAliasesTable,
		Indexes: map[string]*memdb.IndexSchema{
			"id": &memdb.IndexSchema{
				Name:   "id",
				Unique: true,
				Indexer: &memdb.StringFieldIndex{
					Field: "ID",
				},
			},
			"factors": &memdb.IndexSchema{
				Name:   "factors",
				Unique: true,
				Indexer: &memdb.CompoundIndex{
					Indexes: []memdb.Indexer{
						&memdb.StringFieldIndex{
							Field: "MountAccessor",
						},
						&memdb.StringFieldIndex{
							Field: "Name",
						},
					},
				},
			},
		},
	}
}
