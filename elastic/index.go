package elastic

import (
	"context"
	"strconv"
	"strings"
)

var indexMapping = `
{
	"settings": {
		"index.mapping.single_type": true,
		"analysis": {
			"filter": {
				"my_stemmer": {
					"name": "light_english",
					"type": "stemmer"
				},
				"code": {
					"type": "pattern_capture",
					"preserve_original": "true",
					"patterns": [
						"(\\p{Ll}+|\\p{Lu}\\p{Ll}+|\\p{Lu}+)",
						"(\\d+)",
						"(?=([\\p{Lu}]+[\\p{L}]+))",
						"\"((?:\\\"|[^\"]|\\\")*)\"",
						"'((?:\\'|[^']|\\')*)'",
						"\\.([^.]+)(?=\\.|\\s|\\Z)",
						"\\/?([^\\/]+)(?=\\/|\\b)"
					]
				},
				"edgeNGram_filter": {
					"type": "edgeNGram",
					"min_gram": "2",
					"max_gram": "40"
				}
			},
			"analyzer": {
				"default": {
					"filter": [
						"standard",
						"lowercase",
						"my_stemmer"
					],
					"tokenizer": "standard"
				},
				"code_search_analyzer": {
					"filter": [
						"lowercase",
						"asciifolding"
					],
					"type": "custom",
					"tokenizer": "whitespace"
				},
				"path_analyzer": {
					"filter": [
						"lowercase",
						"asciifolding"
					],
					"type": "custom",
					"tokenizer": "path_tokenizer"
				},
				"sha_analyzer": {
					"filter": [
						"lowercase",
						"asciifolding"
					],
					"type": "custom",
					"tokenizer": "sha_tokenizer"
				},
				"code_analyzer": {
					"filter": [
						"code",
						"lowercase",
						"asciifolding",
						"edgeNGram_filter"
					],
					"type": "custom",
					"tokenizer": "whitespace"
				},
				"my_ngram_analyzer": {
					"filter": [
						"lowercase"
					],
					"tokenizer": "my_ngram_tokenizer"
				}
			},
			"tokenizer": {
				"my_ngram_tokenizer": {
					"token_chars": [
						"letter",
						"digit"
					],
					"min_gram": "2",
					"type": "nGram",
					"max_gram": "3"
				},
				"sha_tokenizer": {
					"token_chars": [
						"letter",
						"digit"
					],
					"min_gram": "5",
					"type": "edgeNGram",
					"max_gram": "40"
				},
				"path_tokenizer": {
					"reverse": "true",
					"type": "path_hierarchy"
				}
			}
		}
	},
	"mappings": {
		"doc": {
			"dynamic": "strict",
			"_routing": {
				"required": true
			},
			"properties": {
				"archived": {
						"type": "boolean"
				},
				"assignee_id": {
						"type": "integer"
				},
				"author_id": {
						"type": "integer"
				},
				"blob": {
						"properties": {
								"commit_sha": {
										"analyzer": "sha_analyzer",
										"index_options": "offsets",
										"type": "text"
								},
								"content": {
										"analyzer": "code_analyzer",
										"index_options": "offsets",
										"search_analyzer": "code_search_analyzer",
										"type": "text"
								},
								"file_name": {
										"analyzer": "code_analyzer",
										"search_analyzer": "code_search_analyzer",
										"type": "text"
								},
								"id": {
										"analyzer": "sha_analyzer",
										"index_options": "offsets",
										"type": "text"
								},
								"language": {
										"type": "keyword"
								},
								"oid": {
										"analyzer": "sha_analyzer",
										"index_options": "offsets",
										"type": "text"
								},
								"path": {
										"analyzer": "path_analyzer",
										"type": "text"
								},
								"rid": {
										"type": "keyword"
								},
								"type": {
									"type": "keyword"
								}
						}
				},
				"commit": {
						"properties": {
								"author": {
										"properties": {
												"email": {
														"index_options": "offsets",
														"type": "text"
												},
												"name": {
														"index_options": "offsets",
														"type": "text"
												},
												"time": {
														"format": "basic_date_time_no_millis",
														"type": "date"
												}
										}
								},
								"committer": {
										"properties": {
												"email": {
														"index_options": "offsets",
														"type": "text"
												},
												"name": {
														"index_options": "offsets",
														"type": "text"
												},
												"time": {
														"format": "basic_date_time_no_millis",
														"type": "date"
												}
										}
								},
								"id": {
										"analyzer": "sha_analyzer",
										"index_options": "offsets",
										"type": "text"
								},
								"message": {
										"index_options": "offsets",
										"type": "text"
								},
								"rid": {
										"type": "keyword"
								},
								"sha": {
										"analyzer": "sha_analyzer",
										"index_options": "offsets",
										"type": "text"
								},
								"type": {
									"type": "keyword"
								}
						}
				},
				"confidential": {
						"type": "boolean"
				},
				"content": {
						"index_options": "offsets",
						"type": "text"
				},
				"created_at": {
						"type": "date"
				},
				"description": {
						"index_options": "offsets",
						"type": "text"
				},
				"file_name": {
						"index_options": "offsets",
						"type": "text"
				},
				"id": {
						"type": "integer"
				},
				"iid": {
						"type": "integer"
				},
				"issue": {
						"properties": {
								"assignee_id": {
										"type": "integer"
								},
								"author_id": {
										"type": "integer"
								},
								"confidential": {
										"type": "boolean"
								}
						}
				},
				"issues_access_level": {
						"type": "integer"
				},
				"join_field": {
						"eager_global_ordinals": true,
						"relations": {
								"project": [
										"note",
										"blob",
										"issue",
										"milestone",
										"wiki_blob",
										"commit",
										"merge_request"
								]
						},
						"type": "join"
				},
				"last_activity_at": {
						"type": "date"
				},
				"last_pushed_at": {
						"type": "date"
				},
				"merge_requests_access_level": {
						"type": "integer"
				},
				"merge_status": {
						"type": "text"
				},
				"name": {
						"index_options": "offsets",
						"type": "text"
				},
				"name_with_namespace": {
						"analyzer": "my_ngram_analyzer",
						"index_options": "offsets",
						"type": "text"
				},
				"namespace_id": {
						"type": "integer"
				},
				"note": {
						"index_options": "offsets",
						"type": "text"
				},
				"noteable_id": {
						"type": "keyword"
				},
				"noteable_type": {
						"type": "keyword"
				},
				"path": {
						"index_options": "offsets",
						"type": "text"
				},
				"path_with_namespace": {
						"index_options": "offsets",
						"type": "text"
				},
				"project_id": {
						"type": "integer"
				},
				"repository_access_level": {
						"type": "integer"
				},
				"snippets_access_level": {
						"type": "integer"
				},
				"source_branch": {
						"index_options": "offsets",
						"type": "text"
				},
				"source_project_id": {
						"type": "integer"
				},
				"state": {
						"type": "text"
				},
				"target_branch": {
						"index_options": "offsets",
						"type": "text"
				},
				"target_project_id": {
						"type": "integer"
				},
				"title": {
						"index_options": "offsets",
						"type": "text"
				},
				"type": {
						"type": "keyword"
				},
				"updated_at": {
						"type": "date"
				},
				"visibility_level": {
						"type": "integer"
				},
				"wiki_access_level": {
						"type": "integer"
				}
			}
		}
	}
}
`

// CreateIndex creates an index matching that created by gitlab-elasticsearch-git v1.1.1
func (c *Client) CreateIndex() error {
	info, err := c.Client.NodesInfo().Do(context.Background())
	if err != nil {
		return err
	}

	for _, node := range info.Nodes {
		// Grab the first character of the version string and turn it into an int
		version, _ := strconv.Atoi(string(node.Version[0]))
		if version >= 6 {
			// single_type is an option only available on 5.6 that ES6 cannot handle
			indexMapping = strings.Replace(indexMapping, "\"index.mapping.single_type\": true,", "", 1)
		}

		// We only look at the first node and assume they're all the same version
		break
	}

	createIndex, err := c.Client.CreateIndex(c.IndexName).BodyString(indexMapping).Do(context.Background())
	if err != nil {
		return err
	}

	if !createIndex.Acknowledged {
		return timeoutError
	}

	return nil
}

func (c *Client) DeleteIndex() error {
	deleteIndex, err := c.Client.DeleteIndex(c.IndexName).Do(context.Background())
	if err != nil {
		return err
	}

	if !deleteIndex.Acknowledged {
		return timeoutError
	}

	return nil
}
