package resolvers

import (
	"fmt"

	"github.com/cave/cmd/graphql/schemas"
	"github.com/cave/cmd/models"

	"github.com/cave/pkg/auth"
	"github.com/cave/pkg/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/jinzhu/gorm"
)

var (
	authenticator *auth.Authenticator
)

// ApplyResolvers applies root queries to graphql server
func ApplyResolvers(r *gin.Engine, db *gorm.DB, auth *auth.Authenticator) {
	models.SetRepoDB(db)
	authenticator = auth

	var rootQuery = graphql.NewObject(
		graphql.ObjectConfig{
			Name:        "Query",
			Description: "User type query",
			Fields: graphql.Fields{
				"user": &graphql.Field{
					Type: schemas.UserSchema,
					Args: graphql.FieldConfigArgument{
						"id": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: GetUser,
				},
				"signin": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"email": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"password": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: SignIn,
				},
			},
		})
	var rootMutation = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				/* Signup user
				 */
				"signup": &graphql.Field{
					Type:        graphql.String,
					Args:        schemas.CreateUserSchema,
					Description: "Register new user",
					Resolve:     SignUp,
				},
			},
		})

	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    rootQuery,
			Mutation: rootMutation,
		},
	)

	r.POST("/graphql", middlewares.JWTAuthMiddleware(authenticator), func(ctx *gin.Context) {
		var query struct {
			Query string
		}
		ctx.BindJSON(&query)
		result := executeQuery(query.Query, schema, ctx)
		ctx.JSON(200, result)
	})
}

func executeQuery(query string, schema graphql.Schema, ctx *gin.Context) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       ctx,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %+v", result.Errors)
	}
	return result
}
