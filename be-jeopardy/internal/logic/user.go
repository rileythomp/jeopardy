package logic

import (
	"context"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

var supabase *db.SupabaseDB

func init() {
	var err error
	supabase, err = db.NewSupabaseDB(context.Background())
	if err != nil {
		log.Fatalf("error creating supabase db: %v", err)
	}
}

func GetUserByName(ctx context.Context, name string) (db.User, error) {
	user, err := supabase.GetUserByName(ctx, name)
	if err != nil {
		log.Errorf("error getting user by name: %v", err)
		return db.User{}, err
	}
	return user, nil
}
