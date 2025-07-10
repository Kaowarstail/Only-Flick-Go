package seed

import (
	"log"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
	"golang.org/x/crypto/bcrypt"
)

func Run() {
	db := database.GetDB()

	log.Println("🌱 Début du seed de la base de données...")

	// Helper function to hash passwords
	hashPassword := func(password string) string {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Erreur lors du hachage du mot de passe:", err)
		}
		return string(hashed)
	}

	// --- Utilisateurs ---
	users := []models.User{
		{
			ID: "admin-1", Username: "admin", Email: "admin@onlyflick.com", Password: hashPassword("admin123"),
			Role: models.RoleAdmin, FirstName: "Admin", LastName: "OnlyFlick",
			Biography:      "Administrateur principal de la plateforme OnlyFlick 👑",
			ProfilePicture: "https://i.pravatar.cc/300?img=35", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-1", Username: "alice_photographer", Email: "alice@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "Alice", LastName: "Johnson",
			Biography:      "Professional photographer and visual artist. Capturing life's beautiful moments 📸✨",
			ProfilePicture: "https://i.pravatar.cc/300?img=1", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-2", Username: "bob_fitness", Email: "bob@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "Bob", LastName: "Wilson",
			Biography:      "Fitness coach & nutrition expert. Transform your body and mind 💪🏋️‍♂️",
			ProfilePicture: "https://i.pravatar.cc/300?img=11", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-3", Username: "carol_chef", Email: "carol@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "Carol", LastName: "Martinez",
			Biography:      "Chef & food stylist. Bringing delicious recipes to your kitchen 👩‍🍳🍽️",
			ProfilePicture: "https://i.pravatar.cc/300?img=5", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-4", Username: "david_traveler", Email: "david@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "David", LastName: "Chen",
			Biography:      "World traveler & adventure seeker. Join me on epic journeys ✈️🌍",
			ProfilePicture: "https://i.pravatar.cc/300?img=33", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-5", Username: "emma_artist", Email: "emma@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "Emma", LastName: "Davis",
			Biography:      "Digital artist & illustrator. Creating magic through pixels 🎨✨",
			ProfilePicture: "https://i.pravatar.cc/300?img=9", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-6", Username: "frank_musician", Email: "frank@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "Frank", LastName: "Thompson",
			Biography:      "Musician & producer. Sharing the rhythm of life 🎵🎹",
			ProfilePicture: "https://i.pravatar.cc/300?img=12", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-7", Username: "grace_stylist", Email: "grace@example.com", Password: hashPassword("password123"),
			Role: models.RoleCreator, FirstName: "Grace", LastName: "Kim",
			Biography:      "Fashion stylist & trend setter. Style is a way to say who you are 👗💄",
			ProfilePicture: "https://i.pravatar.cc/300?img=20", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-8", Username: "henry_tech", Email: "henry@example.com", Password: hashPassword("password123"),
			FirstName: "Henry", LastName: "Rodriguez",
			Biography:      "Tech enthusiast and early adopter. Always exploring the latest innovations 💻📱",
			ProfilePicture: "https://i.pravatar.cc/300?img=15", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-9", Username: "irene_bookworm", Email: "irene@example.com", Password: hashPassword("password123"),
			FirstName: "Irene", LastName: "Brown",
			Biography:      "Book lover and literature enthusiast. Lost in the world of stories 📚✍️",
			ProfilePicture: "https://i.pravatar.cc/300?img=16", IsActive: true, IsEmailVerified: true,
		},
		{
			ID: "user-10", Username: "jack_gamer", Email: "jack@example.com", Password: hashPassword("password123"),
			FirstName: "Jack", LastName: "Taylor",
			Biography:      "Professional gamer and streamer. Level up with me! 🎮🕹️",
			ProfilePicture: "https://i.pravatar.cc/300?img=13", IsActive: true, IsEmailVerified: true,
		},
	}
	db.Create(&users)

	// --- Profils créateurs ---
	profiles := []models.CreatorProfile{
		{
			UserID:      "user-1",
			BannerImage: "https://images.unsplash.com/photo-1492691527719-9d1e07e534b4?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://alice-photography.com",
			SocialLinks: `{"instagram": "@alice_photographer", "twitter": "@alicephoto", "youtube": "AlicePhotography"}`,
		},
		{
			UserID:      "user-2",
			BannerImage: "https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://bobfitness.com",
			SocialLinks: `{"instagram": "@bob_fitness", "twitter": "@bobfitpro", "tiktok": "@bobfitness"}`,
		},
		{
			UserID:      "user-3",
			BannerImage: "https://images.unsplash.com/photo-1556909114-f6e7ad7d3136?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://carolscooking.com",
			SocialLinks: `{"instagram": "@carol_chef", "youtube": "CarolsCooking", "pinterest": "carolchef"}`,
		},
		{
			UserID:      "user-4",
			BannerImage: "https://images.unsplash.com/photo-1488646953014-85cb44e25828?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://davidtravels.blog",
			SocialLinks: `{"instagram": "@david_traveler", "youtube": "DavidTravels", "blog": "davidtravels.blog"}`,
		},
		{
			UserID:      "user-5",
			BannerImage: "https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://emmaart.studio",
			SocialLinks: `{"instagram": "@emma_artist", "behance": "emmaartist", "dribbble": "emma_designs"}`,
		},
		{
			UserID:      "user-6",
			BannerImage: "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://frankmusic.com",
			SocialLinks: `{"instagram": "@frank_musician", "spotify": "FrankThompson", "soundcloud": "frankmusic"}`,
		},
		{
			UserID:      "user-7",
			BannerImage: "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1200&h=400&fit=crop",
			WebsiteURL:  "https://gracestyle.com",
			SocialLinks: `{"instagram": "@grace_stylist", "pinterest": "gracestyle", "tiktok": "@gracefashion"}`,
		},
	}
	db.Create(&profiles)

	// --- Plans d'abonnement ---
	plans := []models.SubscriptionPlan{
		// Alice Photography plans
		{CreatorID: "user-1", Name: "Photo Enthusiast", Price: 4.99, Duration: 30, Description: "Access to basic photo tutorials and tips"},
		{CreatorID: "user-1", Name: "Pro Photographer", Price: 12.99, Duration: 30, Description: "Premium photo workshops and 1-on-1 sessions"},

		// Bob Fitness plans
		{CreatorID: "user-2", Name: "Fitness Starter", Price: 9.99, Duration: 30, Description: "Basic workout plans and nutrition guides"},
		{CreatorID: "user-2", Name: "Elite Training", Price: 19.99, Duration: 30, Description: "Personalized training and meal plans"},

		// Carol Chef plans
		{CreatorID: "user-3", Name: "Home Cook", Price: 7.99, Duration: 30, Description: "Easy recipes and cooking basics"},
		{CreatorID: "user-3", Name: "Culinary Master", Price: 15.99, Duration: 30, Description: "Advanced techniques and exclusive recipes"},

		// David Traveler plans
		{CreatorID: "user-4", Name: "Weekend Explorer", Price: 5.99, Duration: 30, Description: "Local travel tips and hidden gems"},
		{CreatorID: "user-4", Name: "Global Adventurer", Price: 14.99, Duration: 30, Description: "International travel guides and planning"},

		// Emma Artist plans
		{CreatorID: "user-5", Name: "Art Appreciator", Price: 6.99, Duration: 30, Description: "Digital art tutorials and process videos"},
		{CreatorID: "user-5", Name: "Creative Pro", Price: 16.99, Duration: 30, Description: "Advanced techniques and custom artwork"},

		// Frank Musician plans
		{CreatorID: "user-6", Name: "Music Lover", Price: 8.99, Duration: 30, Description: "Behind-the-scenes content and demos"},
		{CreatorID: "user-6", Name: "Producer Access", Price: 18.99, Duration: 30, Description: "Production tutorials and exclusive tracks"},

		// Grace Stylist plans
		{CreatorID: "user-7", Name: "Style Basics", Price: 6.99, Duration: 30, Description: "Fashion tips and styling advice"},
		{CreatorID: "user-7", Name: "Fashion Elite", Price: 13.99, Duration: 30, Description: "Personal styling sessions and trend forecasts"},
	}
	db.Create(&plans)

	// --- Contenus ---
	contents := []models.Content{
		// Alice Photography content
		{
			CreatorID: "user-1", Title: "Golden Hour Portrait Session",
			Description: "Behind the scenes of a magical golden hour photoshoot in the city park. Learn about lighting techniques and camera settings.",
			Type:        "image", MediaURL: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=800&h=600&fit=crop",
			ThumbnailURL: "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 1250,
		},
		{
			CreatorID: "user-1", Title: "Advanced Camera Techniques",
			Description: "Exclusive tutorial on advanced DSLR techniques for premium subscribers only.",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1502920917128-1aa500764cbd?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 890,
		},
		{
			CreatorID: "user-1", Title: "Street Photography Tips",
			Description: "My top 10 tips for capturing authentic street moments. Perfect for beginners!",
			Type:        "image", MediaURL: "https://images.unsplash.com/photo-1449824913935-59a10b8d2000?w=800&h=600&fit=crop",
			ThumbnailURL: "https://images.unsplash.com/photo-1449824913935-59a10b8d2000?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 2100,
		},

		// Bob Fitness content
		{
			CreatorID: "user-2", Title: "30-Minute Full Body Workout",
			Description: "High-intensity workout that targets all muscle groups. No equipment needed!",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_2mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 3200,
		},
		{
			CreatorID: "user-2", Title: "Nutrition Guide for Athletes",
			Description: "Complete nutrition breakdown for optimal performance. Premium members get the full meal plan PDF.",
			Type:        "text", MediaURL: "",
			ThumbnailURL: "https://images.unsplash.com/photo-1490645935967-10de6ba17061?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 1800,
		},
		{
			CreatorID: "user-2", Title: "Morning Stretch Routine",
			Description: "Start your day right with this 10-minute energizing stretch sequence.",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1544367567-0f2fcb009e0b?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 2700,
		},

		// Carol Chef content
		{
			CreatorID: "user-3", Title: "Homemade Pasta from Scratch",
			Description: "Learn to make authentic Italian pasta at home. Recipe and technique included!",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_2mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1551183053-bf91a1d81141?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 4100,
		},
		{
			CreatorID: "user-3", Title: "Secret Sauce Recipes",
			Description: "My collection of signature sauce recipes that make any dish extraordinary. Premium exclusive!",
			Type:        "text", MediaURL: "",
			ThumbnailURL: "https://images.unsplash.com/photo-1565299624946-b28f40a0ca4b?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 1500,
		},
		{
			CreatorID: "user-3", Title: "Quick Weeknight Dinners",
			Description: "5 delicious dinner ideas that take 30 minutes or less to prepare.",
			Type:        "image", MediaURL: "https://images.unsplash.com/photo-1565958011703-44f9829ba187?w=800&h=600&fit=crop",
			ThumbnailURL: "https://images.unsplash.com/photo-1565958011703-44f9829ba187?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 2950,
		},

		// David Traveler content
		{
			CreatorID: "user-4", Title: "Hidden Gems of Tokyo",
			Description: "Discover the secret spots in Tokyo that most tourists never see. Complete with maps and insider tips!",
			Type:        "image", MediaURL: "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?w=800&h=600&fit=crop",
			ThumbnailURL: "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 3800,
		},
		{
			CreatorID: "user-4", Title: "Budget Travel Masterclass",
			Description: "How I travel the world on $50 a day. Exclusive strategies and resources for premium members.",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1488646953014-85cb44e25828?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 2200,
		},

		// Emma Artist content
		{
			CreatorID: "user-5", Title: "Digital Portrait Speedpaint",
			Description: "Watch me create a digital portrait from start to finish in real-time. Includes brush settings!",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_2mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 1950,
		},
		{
			CreatorID: "user-5", Title: "Character Design Workshop",
			Description: "Premium workshop on creating memorable characters. Includes layered PSD files and custom brushes.",
			Type:        "image", MediaURL: "https://images.unsplash.com/photo-1513475382585-d06e58bcb0e0?w=800&h=600&fit=crop",
			ThumbnailURL: "https://images.unsplash.com/photo-1513475382585-d06e58bcb0e0?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 1200,
		},

		// Frank Musician content
		{
			CreatorID: "user-6", Title: "Acoustic Sessions at Home",
			Description: "Intimate acoustic performances of my latest songs, recorded in my home studio.",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 2800,
		},
		{
			CreatorID: "user-6", Title: "Music Production Secrets",
			Description: "Learn the techniques I use to produce chart-topping tracks. DAW project files included for premium members.",
			Type:        "video", MediaURL: "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_2mb.mp4",
			ThumbnailURL: "https://images.unsplash.com/photo-1598488035139-bdbb2231ce04?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 1650,
		},

		// Grace Stylist content
		{
			CreatorID: "user-7", Title: "Fall Fashion Lookbook 2024",
			Description: "15 stunning fall outfits styled by me. Perfect for any occasion and budget!",
			Type:        "image", MediaURL: "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=800&h=600&fit=crop",
			ThumbnailURL: "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=300&h=300&fit=crop",
			IsPremium:    false, ViewCount: 3500,
		},
		{
			CreatorID: "user-7", Title: "Personal Color Analysis",
			Description: "Discover your perfect color palette with this comprehensive guide. Premium members get personalized analysis.",
			Type:        "text", MediaURL: "",
			ThumbnailURL: "https://images.unsplash.com/photo-1560472354-b33ff0c44a43?w=300&h=300&fit=crop",
			IsPremium:    true, ViewCount: 1100,
		},
	}
	db.Create(&contents)

	// --- Abonnements ---
	subs := []models.Subscription{
		{SubscriberID: "user-8", CreatorID: "user-1", PlanID: plans[0].ID, StartDate: time.Now().Add(-15 * 24 * time.Hour), EndDate: time.Now().Add(15 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-9", CreatorID: "user-1", PlanID: plans[1].ID, StartDate: time.Now().Add(-10 * 24 * time.Hour), EndDate: time.Now().Add(20 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-10", CreatorID: "user-2", PlanID: plans[2].ID, StartDate: time.Now().Add(-5 * 24 * time.Hour), EndDate: time.Now().Add(25 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-8", CreatorID: "user-3", PlanID: plans[4].ID, StartDate: time.Now().Add(-20 * 24 * time.Hour), EndDate: time.Now().Add(10 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-9", CreatorID: "user-4", PlanID: plans[6].ID, StartDate: time.Now().Add(-12 * 24 * time.Hour), EndDate: time.Now().Add(18 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-10", CreatorID: "user-5", PlanID: plans[8].ID, StartDate: time.Now().Add(-8 * 24 * time.Hour), EndDate: time.Now().Add(22 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-8", CreatorID: "user-6", PlanID: plans[10].ID, StartDate: time.Now().Add(-25 * 24 * time.Hour), EndDate: time.Now().Add(5 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-9", CreatorID: "user-7", PlanID: plans[12].ID, StartDate: time.Now().Add(-18 * 24 * time.Hour), EndDate: time.Now().Add(12 * 24 * time.Hour), IsActive: true},
	}
	db.Create(&subs)

	// --- Likes ---
	likes := []models.Like{
		// Likes for Alice's content
		{UserID: "user-8", ContentID: contents[0].ID},
		{UserID: "user-9", ContentID: contents[0].ID},
		{UserID: "user-10", ContentID: contents[0].ID},
		{UserID: "user-2", ContentID: contents[0].ID},
		{UserID: "user-3", ContentID: contents[0].ID},
		{UserID: "user-8", ContentID: contents[1].ID},
		{UserID: "user-9", ContentID: contents[1].ID},
		{UserID: "user-8", ContentID: contents[2].ID},
		{UserID: "user-10", ContentID: contents[2].ID},
		{UserID: "user-4", ContentID: contents[2].ID},
		{UserID: "user-5", ContentID: contents[2].ID},

		// Likes for Bob's content
		{UserID: "user-8", ContentID: contents[3].ID},
		{UserID: "user-9", ContentID: contents[3].ID},
		{UserID: "user-10", ContentID: contents[3].ID},
		{UserID: "user-1", ContentID: contents[3].ID},
		{UserID: "user-3", ContentID: contents[3].ID},
		{UserID: "user-4", ContentID: contents[3].ID},
		{UserID: "user-9", ContentID: contents[4].ID},
		{UserID: "user-10", ContentID: contents[4].ID},
		{UserID: "user-8", ContentID: contents[5].ID},
		{UserID: "user-9", ContentID: contents[5].ID},
		{UserID: "user-1", ContentID: contents[5].ID},

		// Likes for Carol's content
		{UserID: "user-8", ContentID: contents[6].ID},
		{UserID: "user-9", ContentID: contents[6].ID},
		{UserID: "user-10", ContentID: contents[6].ID},
		{UserID: "user-1", ContentID: contents[6].ID},
		{UserID: "user-2", ContentID: contents[6].ID},
		{UserID: "user-4", ContentID: contents[6].ID},
		{UserID: "user-5", ContentID: contents[6].ID},
		{UserID: "user-8", ContentID: contents[7].ID},
		{UserID: "user-9", ContentID: contents[8].ID},
		{UserID: "user-10", ContentID: contents[8].ID},
		{UserID: "user-2", ContentID: contents[8].ID},

		// Likes for David's content
		{UserID: "user-8", ContentID: contents[9].ID},
		{UserID: "user-9", ContentID: contents[9].ID},
		{UserID: "user-10", ContentID: contents[9].ID},
		{UserID: "user-1", ContentID: contents[9].ID},
		{UserID: "user-2", ContentID: contents[9].ID},
		{UserID: "user-3", ContentID: contents[9].ID},
		{UserID: "user-9", ContentID: contents[10].ID},
		{UserID: "user-10", ContentID: contents[10].ID},

		// Likes for Emma's content
		{UserID: "user-8", ContentID: contents[11].ID},
		{UserID: "user-9", ContentID: contents[11].ID},
		{UserID: "user-10", ContentID: contents[11].ID},
		{UserID: "user-2", ContentID: contents[11].ID},
		{UserID: "user-8", ContentID: contents[12].ID},

		// Likes for Frank's content
		{UserID: "user-8", ContentID: contents[13].ID},
		{UserID: "user-9", ContentID: contents[13].ID},
		{UserID: "user-10", ContentID: contents[13].ID},
		{UserID: "user-1", ContentID: contents[13].ID},
		{UserID: "user-3", ContentID: contents[13].ID},
		{UserID: "user-8", ContentID: contents[14].ID},
		{UserID: "user-9", ContentID: contents[14].ID},

		// Likes for Grace's content
		{UserID: "user-8", ContentID: contents[15].ID},
		{UserID: "user-9", ContentID: contents[15].ID},
		{UserID: "user-10", ContentID: contents[15].ID},
		{UserID: "user-1", ContentID: contents[15].ID},
		{UserID: "user-2", ContentID: contents[15].ID},
		{UserID: "user-4", ContentID: contents[15].ID},
		{UserID: "user-8", ContentID: contents[16].ID},
	}
	db.Create(&likes)

	// --- Commentaires ---
	comments := []models.Comment{
		// Comments on Alice's photography content
		{UserID: "user-8", ContentID: contents[0].ID, Text: "Absolutely stunning composition! The golden hour lighting is perfect 📸✨"},
		{UserID: "user-9", ContentID: contents[0].ID, Text: "Your work always inspires me to pick up my camera. Beautiful shot!"},
		{UserID: "user-10", ContentID: contents[0].ID, Text: "The depth of field in this image is incredible. What lens did you use?"},
		{UserID: "user-2", ContentID: contents[0].ID, Text: "This gives me so much motivation for my own photography journey!"},
		{UserID: "user-8", ContentID: contents[1].ID, Text: "Thanks for sharing these advanced techniques! Game changer 🙌"},
		{UserID: "user-9", ContentID: contents[1].ID, Text: "The camera settings explanation was so helpful. Can't wait to try this!"},
		{UserID: "user-10", ContentID: contents[2].ID, Text: "Perfect timing! I was just planning a street photography walk this weekend."},
		{UserID: "user-4", ContentID: contents[2].ID, Text: "These tips will definitely help me capture better travel photos!"},

		// Comments on Bob's fitness content
		{UserID: "user-8", ContentID: contents[3].ID, Text: "Just finished this workout and I'm exhausted! Great routine 💪"},
		{UserID: "user-9", ContentID: contents[3].ID, Text: "Love that no equipment is needed. Perfect for my small apartment!"},
		{UserID: "user-10", ContentID: contents[3].ID, Text: "Your form explanations are so clear. Thanks for the detailed instructions!"},
		{UserID: "user-1", ContentID: contents[3].ID, Text: "Adding this to my weekly routine. The intensity is just right!"},
		{UserID: "user-9", ContentID: contents[4].ID, Text: "The nutrition breakdown is exactly what I needed for my training!"},
		{UserID: "user-8", ContentID: contents[5].ID, Text: "This morning routine has become essential for my day. Thank you!"},
		{UserID: "user-1", ContentID: contents[5].ID, Text: "Such a relaxing way to start the morning. My back feels so much better!"},

		// Comments on Carol's cooking content
		{UserID: "user-8", ContentID: contents[6].ID, Text: "Made this last night and it was incredible! My family loved it 🍝"},
		{UserID: "user-9", ContentID: contents[6].ID, Text: "Your pasta technique is amazing. Finally got the texture perfect!"},
		{UserID: "user-10", ContentID: contents[6].ID, Text: "The step-by-step process was so easy to follow. Restaurant quality at home!"},
		{UserID: "user-1", ContentID: contents[6].ID, Text: "This brought back memories of my trip to Italy. Authentic flavor!"},
		{UserID: "user-8", ContentID: contents[7].ID, Text: "These secret sauces have elevated all my dishes. Worth every penny!"},
		{UserID: "user-9", ContentID: contents[8].ID, Text: "Perfect for busy weeknights! All five recipes are now in my rotation."},
		{UserID: "user-2", ContentID: contents[8].ID, Text: "Love how you make healthy eating accessible and delicious!"},

		// Comments on David's travel content
		{UserID: "user-8", ContentID: contents[9].ID, Text: "Tokyo is now on my bucket list! These spots look incredible 🇯🇵"},
		{UserID: "user-9", ContentID: contents[9].ID, Text: "Your local insights are invaluable. Can't wait to explore these places!"},
		{UserID: "user-10", ContentID: contents[9].ID, Text: "The photography in this post is breathtaking. Makes me want to book a flight!"},
		{UserID: "user-2", ContentID: contents[9].ID, Text: "Thanks for including the maps and insider tips. So helpful for planning!"},
		{UserID: "user-9", ContentID: contents[10].ID, Text: "Your budget travel strategies are life-changing. Finally planning my dream trip!"},

		// Comments on Emma's art content
		{UserID: "user-8", ContentID: contents[11].ID, Text: "Watching your process is mesmerizing! The final result is stunning 🎨"},
		{UserID: "user-9", ContentID: contents[11].ID, Text: "Your brush technique explanations are so helpful for my own art journey!"},
		{UserID: "user-10", ContentID: contents[11].ID, Text: "The time-lapse format is perfect. Love seeing art come to life!"},
		{UserID: "user-8", ContentID: contents[12].ID, Text: "The character design principles you shared are pure gold. Thank you!"},

		// Comments on Frank's music content
		{UserID: "user-8", ContentID: contents[13].ID, Text: "Your acoustic sessions always give me chills. Beautiful music! 🎵"},
		{UserID: "user-9", ContentID: contents[13].ID, Text: "The intimacy of these home recordings is incredible. Love the raw emotion!"},
		{UserID: "user-10", ContentID: contents[13].ID, Text: "This song has been on repeat all week. When's the full album coming?"},
		{UserID: "user-1", ContentID: contents[13].ID, Text: "Your voice and guitar work perfectly together. So talented!"},
		{UserID: "user-8", ContentID: contents[14].ID, Text: "The production techniques you shared are game-changing for my music!"},

		// Comments on Grace's fashion content
		{UserID: "user-8", ContentID: contents[15].ID, Text: "Every single look is perfection! Already planning my fall wardrobe 👗"},
		{UserID: "user-9", ContentID: contents[15].ID, Text: "Your styling tips make high fashion accessible. Love the budget options!"},
		{UserID: "user-10", ContentID: contents[15].ID, Text: "The color coordination in look #7 is absolutely stunning!"},
		{UserID: "user-1", ContentID: contents[15].ID, Text: "You have such an eye for putting together outfits. Inspiring!"},
		{UserID: "user-8", ContentID: contents[16].ID, Text: "The personal color analysis was so accurate! Changed my whole wardrobe approach."},
	}
	db.Create(&comments)

	log.Println("✅ Seed terminé avec succès.")
}
