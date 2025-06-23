package seed

import (
	"log"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/database"
	"github.com/Kaowarstail/Only-Flick-Go/models"
)

func Run() {
	db := database.GetDB()

	log.Println("🌱 Début du seed de la base de données...")

	// --- Utilisateurs ---
	users := []models.User{
		{ID: "user-1", Username: "alice", Email: "alice@example.com", Password: "-", Role: models.RoleCreator},
		{ID: "user-2", Username: "bob", Email: "bob@example.com", Password: "-", Role: models.RoleCreator},
		{ID: "user-3", Username: "carol", Email: "carol@example.com", Password: "-", Role: models.RoleCreator},
		{ID: "user-4", Username: "dan", Email: "dan@example.com", Password: "-"},
		{ID: "user-5", Username: "eve", Email: "eve@example.com", Password: "-"},
		{ID: "user-6", Username: "frank", Email: "frank@example.com", Password: "-"},
		{ID: "user-7", Username: "grace", Email: "grace@example.com", Password: "-"},
		{ID: "user-8", Username: "henry", Email: "henry@example.com", Password: "-"},
		{ID: "user-9", Username: "irene", Email: "irene@example.com", Password: "-"},
		{ID: "user-10", Username: "jack", Email: "jack@example.com", Password: "-"},
	}
	db.Create(&users)

	// --- Profils créateurs ---
	profiles := []models.CreatorProfile{
		{UserID: "user-1", BannerImage: "banner1.jpg", WebsiteURL: "https://alice.com"},
		{UserID: "user-2", BannerImage: "banner2.jpg", WebsiteURL: "https://bob.com"},
		{UserID: "user-3", BannerImage: "banner3.jpg", WebsiteURL: "https://carol.com"},
	}
	db.Create(&profiles)

	// --- Plans d'abonnement ---
	plans := []models.SubscriptionPlan{
		{CreatorID: "user-1", Name: "Bronze", Price: 4.99, Duration: 30},
		{CreatorID: "user-1", Name: "Gold", Price: 9.99, Duration: 30},
		{CreatorID: "user-2", Name: "Premium", Price: 14.99, Duration: 30},
		{CreatorID: "user-3", Name: "Basic", Price: 5.99, Duration: 30},
	}
	db.Create(&plans)

	// --- Contenus ---
	contents := []models.Content{
		{CreatorID: "user-1", Title: "Photo shoot", Type: "image", MediaURL: "media1.jpg", IsPremium: false},
		{CreatorID: "user-1", Title: "Behind the scenes", Type: "video", MediaURL: "media2.mp4", IsPremium: true},
		{CreatorID: "user-2", Title: "Fitness tips", Type: "text", MediaURL: "media3.txt", IsPremium: false},
		{CreatorID: "user-3", Title: "Cooking live", Type: "video", MediaURL: "media4.mp4", IsPremium: true},
	}
	db.Create(&contents)

	// --- Abonnements ---
	subs := []models.Subscription{
		{SubscriberID: "user-4", CreatorID: "user-1", PlanID: plans[0].ID, StartDate: time.Now(), EndDate: time.Now().Add(30 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-5", CreatorID: "user-1", PlanID: plans[1].ID, StartDate: time.Now(), EndDate: time.Now().Add(30 * 24 * time.Hour), IsActive: true},
		{SubscriberID: "user-6", CreatorID: "user-2", PlanID: plans[2].ID, StartDate: time.Now(), EndDate: time.Now().Add(30 * 24 * time.Hour), IsActive: true},
	}
	db.Create(&subs)

	// --- Likes ---
	likes := []models.Like{
		{UserID: "user-4", ContentID: contents[0].ID},
		{UserID: "user-5", ContentID: contents[0].ID},
		{UserID: "user-6", ContentID: contents[0].ID},
		{UserID: "user-7", ContentID: contents[1].ID},
		{UserID: "user-8", ContentID: contents[1].ID},
		{UserID: "user-9", ContentID: contents[2].ID},
		{UserID: "user-10", ContentID: contents[3].ID},
		{UserID: "user-4", ContentID: contents[3].ID},
	}
	db.Create(&likes)

	// --- Commentaires ---
	comments := []models.Comment{
		{UserID: "user-4", ContentID: contents[0].ID, Text: "Super contenu !"},
		{UserID: "user-5", ContentID: contents[0].ID, Text: "Magnifique travail !"},
		{UserID: "user-6", ContentID: contents[0].ID, Text: "Très inspirant."},
		{UserID: "user-7", ContentID: contents[1].ID, Text: "Merci pour ce partage !"},
		{UserID: "user-8", ContentID: contents[1].ID, Text: "Excellent !"},
		{UserID: "user-9", ContentID: contents[2].ID, Text: "Très utile, bravo."},
		{UserID: "user-10", ContentID: contents[3].ID, Text: "Hâte de voir la suite."},
		{UserID: "user-4", ContentID: contents[3].ID, Text: "Super live !"},
	}
	db.Create(&comments)

	log.Println("✅ Seed terminé avec succès.")
}
