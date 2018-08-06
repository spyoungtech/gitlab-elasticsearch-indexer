package elastic

import (
	"context"
)

var indexMapping = `
{
	"settings": {
		"analysis": {
			"filter": {
				"my_stemmer": {
					"name": "light_english",
					"type": "stemmer"
				},
				"code": {
					"type": "pattern_capture",
					"preserve_original": "1",
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
		"milestone": {
			"_parent": {
				"type": "project"
			},
			"_routing": {
				"required": true
			},
			"properties": {
				"created_at": {
					"type": "date"
				},
				"description": {
					"type": "text",
					"index_options": "offsets"
				},
				"id": {
					"type": "integer"
				},
				"project_id": {
					"type": "integer"
				},
				"title": {
					"type": "text",
					"index_options": "offsets"
				},
				"updated_at": {
					"type": "date"
				}
			}
		},
		"note": {
			"_parent": {
				"type": "project"
			},
			"_routing": {
				"required": true
			},
			"properties": {
				"created_at": {
					"type": "date"
				},
				"id": {
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
				"note": {
					"type": "text",
					"index_options": "offsets"
				},
				"noteable_id": {
					"type": "integer"
				},
				"noteable_type": {
					"type": "keyword"
				},
				"project_id": {
					"type": "integer"
				},
				"updated_at": {
					"type": "date"
				}
			}
		},
		"project_wiki": {
			"_parent": {
				"type": "project"
			},
			"_routing": {
				"required": true
			},
			"properties": {
				"blob": {
					"properties": {
						"commit_sha": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"content": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "code_analyzer",
							"search_analyzer": "code_search_analyzer"
						},
						"file_name": {
							"type": "text",
							"analyzer": "code_analyzer",
							"search_analyzer": "code_search_analyzer"
						},
						"id": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"language": {
							"type": "keyword"
						},
						"oid": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"path": {
							"type": "text",
							"analyzer": "path_analyzer"
						},
						"rid": {
							"type": "keyword"
						}
					}
				},
				"commit": {
					"properties": {
						"author": {
							"properties": {
								"email": {
									"type": "text",
									"index_options": "offsets"
								},
								"name": {
									"type": "text",
									"index_options": "offsets"
								},
								"time": {
									"type": "date",
									"format": "basic_date_time_no_millis"
								}
							}
						},
						"commiter": {
							"properties": {
								"email": {
									"type": "text",
									"index_options": "offsets"
								},
								"name": {
									"type": "text",
									"index_options": "offsets"
								},
								"time": {
									"type": "date",
									"format": "basic_date_time_no_millis"
								}
							}
						},
						"id": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"message": {
							"type": "text",
							"index_options": "offsets"
						},
						"rid": {
							"type": "keyword"
						},
						"sha": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						}
					}
				}
			}
		},
		"issue": {
			"_parent": {
				"type": "project"
			},
			"_routing": {
				"required": true
			},
			"properties": {
				"assignee_id": {
					"type": "integer"
				},
				"author_id": {
					"type": "integer"
				},
				"confidential": {
					"type": "boolean"
				},
				"created_at": {
					"type": "date"
				},
				"description": {
					"type": "text",
					"index_options": "offsets"
				},
				"id": {
					"type": "integer"
				},
				"iid": {
					"type": "integer"
				},
				"project_id": {
					"type": "integer"
				},
				"state": {
					"type": "text"
				},
				"title": {
					"type": "text",
					"index_options": "offsets"
				},
				"updated_at": {
					"type": "date"
				}
			}
		},
		"merge_request": {
			"_parent": {
				"type": "project"
			},
			"_routing": {
				"required": true
			},
			"properties": {
				"author_id": {
					"type": "integer"
				},
				"created_at": {
					"type": "date"
				},
				"description": {
					"type": "text",
					"index_options": "offsets"
				},
				"id": {
					"type": "integer"
				},
				"iid": {
					"type": "integer"
				},
				"merge_status": {
					"type": "text"
				},
				"source_branch": {
					"type": "text",
					"index_options": "offsets"
				},
				"source_project_id": {
					"type": "integer"
				},
				"state": {
					"type": "text"
				},
				"target_branch": {
					"type": "text",
					"index_options": "offsets"
				},
				"target_project_id": {
					"type": "integer"
				},
				"title": {
					"type": "text",
					"index_options": "offsets"
				},
				"updated_at": {
					"type": "date"
				}
			}
		},
		"repository": {
			"_parent": {
				"type": "project"
			},
			"_routing": {
				"required": true
			},
			"properties": {
				"blob": {
					"properties": {
						"commit_sha": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"content": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "code_analyzer",
							"search_analyzer": "code_search_analyzer"
						},
						"file_name": {
							"type": "text",
							"analyzer": "code_analyzer",
							"search_analyzer": "code_search_analyzer"
						},
						"id": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"language": {
							"type": "keyword"
						},
						"oid": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"path": {
							"type": "text",
							"analyzer": "path_analyzer"
						},
						"rid": {
							"type": "keyword"
						},
						"type": {
							"type": "text",
							"fields": {
								"keyword": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						}
					}
				},
				"commit": {
					"properties": {
						"author": {
							"properties": {
								"email": {
									"type": "text",
									"index_options": "offsets"
								},
								"name": {
									"type": "text",
									"index_options": "offsets"
								},
								"time": {
									"type": "date",
									"format": "basic_date_time_no_millis"
								}
							}
						},
						"commiter": {
							"properties": {
								"email": {
									"type": "text",
									"index_options": "offsets"
								},
								"name": {
									"type": "text",
									"index_options": "offsets"
								},
								"time": {
									"type": "date",
									"format": "basic_date_time_no_millis"
								}
							}
						},
						"committer": {
							"properties": {
								"email": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"name": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"time": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								}
							}
						},
						"id": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"message": {
							"type": "text",
							"index_options": "offsets"
						},
						"rid": {
							"type": "keyword"
						},
						"sha": {
							"type": "text",
							"index_options": "offsets",
							"analyzer": "sha_analyzer"
						},
						"type": {
							"type": "text",
							"fields": {
								"keyword": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						}
					}
				}
			}
		},
		"project": {
			"properties": {
				"archived": {
					"type": "boolean"
				},
				"created_at": {
					"type": "date"
				},
				"description": {
					"type": "text",
					"index_options": "offsets"
				},
				"id": {
					"type": "integer"
				},
				"issues_access_level": {
					"type": "integer"
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
				"name": {
					"type": "text",
					"index_options": "offsets"
				},
				"name_with_namespace": {
					"type": "text",
					"index_options": "offsets",
					"analyzer": "my_ngram_analyzer"
				},
				"namespace_id": {
					"type": "integer"
				},
				"path": {
					"type": "text",
					"index_options": "offsets"
				},
				"path_with_namespace": {
					"type": "text",
					"index_options": "offsets"
				},
				"repository_access_level": {
					"type": "integer"
				},
				"snippets_access_level": {
					"type": "integer"
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
		},
		"snippet": {
			"properties": {
				"author_id": {
					"type": "integer"
				},
				"content": {
					"type": "text",
					"index_options": "offsets"
				},
				"created_at": {
					"type": "date"
				},
				"file_name": {
					"type": "text",
					"index_options": "offsets"
				},
				"id": {
					"type": "integer"
				},
				"project_id": {
					"type": "integer"
				},
				"state": {
					"type": "text"
				},
				"title": {
					"type": "text",
					"index_options": "offsets"
				},
				"updated_at": {
					"type": "date"
				},
				"visibility_level": {
					"type": "integer"
				}
			}
		}
	}
}
`

// CreateIndex creates an index matching that created by gitlab-elasticsearch-git v1.1.1
func (c *Client) CreateIndex() error {
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
