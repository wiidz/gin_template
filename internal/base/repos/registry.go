package repos

import (
	"log"

	"github.com/wiidz/gin_template/internal/domain/shared/user/entity"

	"github.com/wiidz/goutil/mngs/psqlMng"
	repoMng "github.com/wiidz/goutil/mngs/repoMng"
)

// M is the global repository manager (supports multi-DB via repoMng if needed).
var M *repoMng.Manager

// User is the concrete implementation of the user repository interface.
var User = struct {
	Repo *repoMng.Repo[entity.UserEntity]
}{}

// Setup initializes the global manager and entity repositories.
func Setup(psql *psqlMng.Manager) {
	if M == nil {
		M = repoMng.NewManager()
	}
	if psql == nil {
		log.Printf("repos: postgres manager nil, skip setup")
		return
	}

	M.SetupDefault(psql.DB())

	// initialize entity repos on default DB
	User.Repo = repoMng.RepoOf[entity.UserEntity](M.Default().DB())
}
