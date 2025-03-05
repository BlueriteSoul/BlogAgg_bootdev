package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BlueriteSoul/BlogAgg_bootdev/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		errStr := fmt.Sprintf("User doesn't exist: %s", name)
		return errors.New(errStr)
	}
	err1 := s.cfg.SetUser(name)
	if err1 != nil {
		errStr := fmt.Sprintf("couldn't set the user as current in the .gatorconfig.json: %s", name)
		return errors.New(errStr)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]
	newUsr := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name}
	_, err := s.db.GetUser(context.Background(), newUsr.Name)
	if err != nil {
		s.db.CreateUser(context.Background(), newUsr)
		fmt.Println("User created: ", newUsr.Name)
		err := s.cfg.SetUser(newUsr.Name)
		if err != nil {
			errStr := fmt.Sprintf("couldn't set the user as current in the .gatorconfig.json: %s", newUsr.Name)
			return errors.New(errStr)
		}
		return nil
	}
	fmt.Println("User already exists")
	return errors.New("User already exists")

}
func handlerReset(s *state, cmd command) error {
	err := s.db.DropAllUsers(context.Background())
	if err != nil {
		return errors.New("Error occured while resetting the database")
	}
	fmt.Printf("All rows in users table are gone\n")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return errors.New("Error occured while getting the users from the database")
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}
	return nil
}
func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time-between-res> e.g. 1s/1m/1h", cmd.Name)
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("couldn't parse time: %w", err)
	}
	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}
	name := cmd.Args[0]
	URL := cmd.Args[1]

	newFeed := database.AddFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name, Url: URL, UserID: user.ID}
	dbFeedReturned, err := s.db.AddFeed(context.Background(), newFeed)
	if err != nil {
		return fmt.Errorf("couln't create new feed, error: %w", err)
	}
	followedFeeds := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: dbFeedReturned.ID}
	dbFeedFollowReturned, err := s.db.CreateFeedFollow(context.Background(), followedFeeds)
	if err != nil {
		return fmt.Errorf("couln't update feed_follow, error: %w", err)
	}
	fmt.Println(dbFeedReturned)
	fmt.Println(dbFeedFollowReturned)
	return nil
}
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couln't get feeds from the DB, error: %w", err)
	}
	fmt.Println(feeds)
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Couldn't query a feed, %w", err)
	}
	newFeedToFollow := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID}
	followedFeed, err := s.db.CreateFeedFollow(context.Background(), newFeedToFollow)
	if err != nil {
		return fmt.Errorf("couldn't follow a feed, error: %w", err)
	}
	fmt.Println(followedFeed)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("Couldn't query a user, %w", err)
	}
	fmt.Printf("%+v\n", followedFeeds)
	for _, fF := range followedFeeds {
		fmt.Println(fF.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	toUnfollow := database.UnfollowParams{user.Name, cmd.Args[0]}
	err := s.db.Unfollow(context.Background(), toUnfollow)
	if err != nil {
		return fmt.Errorf("Couldn't unfollow, %w", err)
	}

	fmt.Println("Unfollowed")
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32
	if len(cmd.Args) == 0 {
		limit = 2
	} else {
		temp, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("argument must be an integer, %w", err)
		}
		limit = int32(temp)
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{Name: user.Name, Limit: limit})
	if err != nil {
		return fmt.Errorf("Couldn't get posts, %w", err)
	}
	for _, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.Description)
	}
	return nil
}
