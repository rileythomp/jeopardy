package supabase

import (
	"context"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

func GetUserByName(ctx context.Context, name string) (any, error) {
	db, err := db.NewSupabaseDB()
	if err != nil {
		log.Errorf("error creating supabase db: %v", err)
		return nil, err
	}
	defer db.Close()
	user, err := db.GetUserByName(ctx, name)
	if err != nil {
		log.Errorf("error getting user by name: %v", err)
		return nil, err
	}
	return user, nil
}
