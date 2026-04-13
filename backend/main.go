		&models.User{},
		&models.Feed{},
		&models.Article{},
		&models.Tag{},
		// AutoMigrate database schema
	if err := database.AutoMigrate(
		&models.User{},
		&models.Feed{},
		&models.Article{},
		&models.Tag{},
		&models.PushConfig{},
		&models.PushLog{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	log.Println("Database schema migrated successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	feedRepo := repository.NewFeedRepository(database)
	articleRepo := repository.NewArticleRepository(database)
	tagRepo := repository.NewTagRepository(database)

