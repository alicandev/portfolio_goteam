package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	lgcBoardAPI "github.com/kxplxn/goteam/internal/board"
	taskAPI "github.com/kxplxn/goteam/internal/task/task"
	tasksAPI "github.com/kxplxn/goteam/internal/task/tasks"
	teamAPI "github.com/kxplxn/goteam/internal/team"
	boardAPI "github.com/kxplxn/goteam/internal/team/board"
	loginAPI "github.com/kxplxn/goteam/internal/user/login"
	registerAPI "github.com/kxplxn/goteam/internal/user/register"
	"github.com/kxplxn/goteam/pkg/api"
	"github.com/kxplxn/goteam/pkg/auth"
	"github.com/kxplxn/goteam/pkg/db/tasktable"
	"github.com/kxplxn/goteam/pkg/db/teamtable"
	"github.com/kxplxn/goteam/pkg/db/usertable"
	lgcBoardTable "github.com/kxplxn/goteam/pkg/legacydb/board"
	lgcUserTable "github.com/kxplxn/goteam/pkg/legacydb/user"
	pkgLog "github.com/kxplxn/goteam/pkg/log"
	"github.com/kxplxn/goteam/pkg/token"
)

func main() {
	// Create a logger for the app.
	log := pkgLog.New()

	// Load environment variables from .env file.
	err := godotenv.Load()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Ensure that the necessary env vars were set.
	env := api.NewEnv()
	if err := env.Validate(); err != nil {
		log.Fatal(err.Error())
		os.Exit(2)
	}

	cfgDynamoDB, err := config.LoadDefaultConfig(
		context.Background(), config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	svcDynamo := dynamodb.NewFromConfig(cfgDynamoDB)

	// Create dependencies that are used by multiple handlers.
	db, err := sql.Open("postgres", env.DBConnStr)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(3)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err.Error())
		os.Exit(4)
	}
	jwtValidator := auth.NewJWTValidator(env.JWTKey)
	userSelector := lgcUserTable.NewSelector(db)

	// Register handlers for API routes.
	mux := http.NewServeMux()

	mux.Handle("/user/register", api.NewHandler(
		nil, map[string]api.MethodHandler{
			http.MethodPost: registerAPI.NewPostHandler(
				registerAPI.NewUserValidator(
					registerAPI.NewUsernameValidator(),
					registerAPI.NewPasswordValidator(),
				),
				token.DecodeInvite,
				registerAPI.NewPasswordHasher(),
				usertable.NewInserter(svcDynamo),
				token.EncodeAuth,
				log,
			),
		},
	))

	mux.Handle("/user/login", api.NewHandler(nil, map[string]api.MethodHandler{
		http.MethodPost: loginAPI.NewPostHandler(
			loginAPI.NewValidator(),
			usertable.NewRetriever(svcDynamo),
			loginAPI.NewPasswordComparator(),
			token.EncodeAuth,
			log,
		),
	}))

	mux.Handle("/team", api.NewHandler(nil, map[string]api.MethodHandler{
		http.MethodGet: teamAPI.NewGetHandler(
			token.DecodeAuth,
			teamtable.NewRetriever(svcDynamo),
			teamtable.NewInserter(svcDynamo),
			log,
		),
	}))

	mux.Handle("/team/board", api.NewHandler(nil, map[string]api.MethodHandler{
		http.MethodDelete: boardAPI.NewDeleteHandler(
			token.DecodeAuth,
			token.DecodeState,
			teamtable.NewBoardDeleter(svcDynamo),
			log,
		),
		http.MethodPatch: boardAPI.NewPatchHandler(
			token.DecodeAuth,
			token.DecodeState,
			boardAPI.NewIDValidator(),
			boardAPI.NewNameValidator(),
			teamtable.NewBoardUpdater(svcDynamo),
			log,
		),
	}))

	// TODO: remove once fully migrated to DynamoDB
	mux.Handle("/board", api.NewHandler(
		jwtValidator, map[string]api.MethodHandler{
			http.MethodPost: lgcBoardAPI.NewPOSTHandler(
				userSelector,
				lgcBoardAPI.NewNameValidator(),
				lgcBoardTable.NewCounter(db),
				lgcBoardTable.NewInserter(db),
				log,
			),
		},
	))

	taskTitleValidator := taskAPI.NewTitleValidator()
	mux.Handle("/task", api.NewHandler(
		jwtValidator, map[string]api.MethodHandler{
			http.MethodPost: taskAPI.NewPostHandler(
				token.DecodeAuth,
				token.DecodeState,
				taskTitleValidator,
				taskTitleValidator,
				taskAPI.NewColNoValidator(),
				tasktable.NewInserter(svcDynamo),
				token.EncodeState,
				log,
			),
			http.MethodPatch: taskAPI.NewPatchHandler(
				token.DecodeAuth,
				token.DecodeState,
				taskTitleValidator,
				taskTitleValidator,
				tasktable.NewUpdater(svcDynamo),
				log,
			),
			http.MethodDelete: taskAPI.NewDeleteHandler(
				token.DecodeAuth,
				token.DecodeState,
				tasktable.NewDeleter(svcDynamo),
				token.EncodeState,
				log,
			),
		},
	))

	mux.Handle("/tasks", api.NewHandler(
		jwtValidator, map[string]api.MethodHandler{
			http.MethodPatch: tasksAPI.NewPatchHandler(
				token.DecodeAuth,
				token.DecodeState,
				tasksAPI.NewColNoValidator(),
				tasktable.NewMultiUpdater(svcDynamo),
				token.EncodeState,
				log,
			),
			http.MethodGet: tasksAPI.NewGetHandler(
				token.DecodeAuth,
				tasktable.NewMultiRetriever(svcDynamo),
				log,
			),
		},
	))

	// Serve the app using the ServeMux.
	log.Info("running server at port " + env.Port)
	if err := http.ListenAndServe(":"+env.Port, mux); err != nil {
		log.Fatal(err.Error())
		os.Exit(5)
	}
}
