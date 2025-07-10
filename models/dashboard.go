package models

import (
	"time"
)

// DashboardStats représente les statistiques générales de la plateforme
type DashboardStats struct {
	TotalUsers       int     `json:"total_users"`
	TotalCreators    int     `json:"total_creators"`
	TotalSubscribers int     `json:"total_subscribers"`
	TotalRevenue     float64 `json:"total_revenue"`
	MonthlyRevenue   float64 `json:"monthly_revenue"`
	WeeklyRevenue    float64 `json:"weekly_revenue"`
	TotalContents    int     `json:"total_contents"`
	NewUsersWeek     int     `json:"new_users_week"`
	NewUsersMonth    int     `json:"new_users_month"`
	PendingReports   int     `json:"pending_reports"`
}

// RevenueStats représente les statistiques de revenus par période
type RevenueStats struct {
	Period string  `json:"period"` // daily, weekly, monthly, yearly
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

// UserGrowthStats représente la croissance des utilisateurs
type UserGrowthStats struct {
	Date        string `json:"date"`
	NewUsers    int    `json:"new_users"`
	TotalUsers  int    `json:"total_users"`
	NewCreators int    `json:"new_creators"`
}

// ContentStats représente les statistiques de contenu
type ContentStats struct {
	TotalContents   int `json:"total_contents"`
	FreeContents    int `json:"free_contents"`
	PremiumContents int `json:"premium_contents"`
	ContentsToday   int `json:"contents_today"`
	ContentsWeek    int `json:"contents_week"`
	ContentsMonth   int `json:"contents_month"`
}

// ReportStats représente les statistiques de signalements
type ReportStats struct {
	TotalReports    int `json:"total_reports"`
	PendingReports  int `json:"pending_reports"`
	ResolvedReports int `json:"resolved_reports"`
	ReportsToday    int `json:"reports_today"`
	ReportsWeek     int `json:"reports_week"`
}

// TopCreatorStats représente les statistiques des top créateurs
type TopCreatorStats struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	ProfilePicture  string  `json:"profile_picture"`
	SubscriberCount int     `json:"subscriber_count"`
	ContentCount    int     `json:"content_count"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`
}

// AdminDashboardData représente toutes les données du dashboard admin
type AdminDashboardData struct {
	Overview     DashboardStats    `json:"overview"`
	RevenueChart []RevenueStats    `json:"revenue_chart"`
	UserGrowth   []UserGrowthStats `json:"user_growth"`
	ContentStats ContentStats      `json:"content_stats"`
	ReportStats  ReportStats       `json:"report_stats"`
	TopCreators  []TopCreatorStats `json:"top_creators"`
	GeneratedAt  time.Time         `json:"generated_at"`
}
