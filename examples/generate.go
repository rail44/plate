//go:generate go run generate.go

package main

import (
	"log"

	"github.com/rail44/plate"
	"github.com/rail44/plate/examples/models"
)

func main() {
	// Define table schemas
	userSchema := plate.TableSchema{
		TableName: "user",
		Model:     models.User{},
	}

	postSchema := plate.TableSchema{
		TableName: "post",
		Model:     models.Post{},
	}

	tagSchema := plate.TableSchema{
		TableName: "tag",
		Model:     models.Tag{},
	}

	postTagSchema := plate.TableSchema{
		TableName: "post_tag",
		Model:     models.PostTag{},
	}

	// Define database schema
	schema := plate.Schema{
		Tables: []plate.TableConfig{
			{
				Schema:    userSchema,
				Relations: []plate.Relation{
					// No BelongsTo relations for User
				},
			},
			{
				Schema: postSchema,
				Relations: []plate.Relation{
					{
						Name:        "Author",
						Target:      "User",
						From:        "UserID",
						To:          "ID",
						ReverseName: "Posts", // This will generate User.Posts()
					},
				},
			},
			{
				Schema:    tagSchema,
				Relations: []plate.Relation{
					// No BelongsTo relations for Tag
				},
			},
		},
		Junctions: []plate.JunctionConfig{
			{
				Schema: postTagSchema,
				Relations: []plate.Relation{
					{
						Name:        "Post",
						Target:      "Post",
						From:        "PostID",
						To:          "ID",
						ReverseName: "Tags", // This will generate Post.Tags()
					},
					{
						Name:        "Tag",
						Target:      "Tag",
						From:        "TagID",
						To:          "ID",
						ReverseName: "Posts", // This will generate Tag.Posts()
					},
				},
			},
		},
	}

	// Generate code (with clean to remove old files)
	gen := plate.NewGenerator()
	opts := plate.GenerateOptions{
		OutputDir: "./generated",
		Clean:     true,
	}
	if err := gen.Generate(schema, opts); err != nil {
		log.Fatal(err)
	}

	log.Println("Code generation completed successfully!")
}
