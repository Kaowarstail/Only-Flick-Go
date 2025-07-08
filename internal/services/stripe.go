package services

import (
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/payout"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/setupintent"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/transfer"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeService gère les interactions avec l'API Stripe
type StripeService struct {
	secretKey string
}

// NewStripeService crée une nouvelle instance du service Stripe
func NewStripeService(secretKey string) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		secretKey: secretKey,
	}
}

// CreateCustomer crée un nouveau client Stripe
func (s *StripeService) CreateCustomer(email, name string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	return customer.New(params)
}

// GetCustomer récupère un client Stripe
func (s *StripeService) GetCustomer(customerID string) (*stripe.Customer, error) {
	return customer.Get(customerID, nil)
}

// CreateSetupIntent crée un SetupIntent pour configurer une méthode de paiement
func (s *StripeService) CreateSetupIntent(customerID string) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Usage: stripe.String("off_session"),
	}

	return setupintent.New(params)
}

// AttachPaymentMethod attache une méthode de paiement à un client
func (s *StripeService) AttachPaymentMethod(paymentMethodID, customerID string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	return paymentmethod.Attach(paymentMethodID, params)
}

// DetachPaymentMethod détache une méthode de paiement d'un client
func (s *StripeService) DetachPaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error) {
	return paymentmethod.Detach(paymentMethodID, nil)
}

// GetPaymentMethod récupère une méthode de paiement
func (s *StripeService) GetPaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error) {
	return paymentmethod.Get(paymentMethodID, nil)
}

// ListPaymentMethods liste les méthodes de paiement d'un client
func (s *StripeService) ListPaymentMethods(customerID string) ([]*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	iter := paymentmethod.List(params)
	var methods []*stripe.PaymentMethod

	for iter.Next() {
		methods = append(methods, iter.PaymentMethod())
	}

	return methods, iter.Err()
}

// CreatePaymentIntent crée un PaymentIntent pour un paiement
func (s *StripeService) CreatePaymentIntent(amount int64, currency, customerID, paymentMethodID string, metadata map[string]string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		ConfirmationMethod: stripe.String("manual"),
		Confirm:            stripe.Bool(true),
		PaymentMethod:      stripe.String(paymentMethodID),
		Metadata:           metadata,
	}

	return paymentintent.New(params)
}

// ConfirmPaymentIntent confirme un PaymentIntent
func (s *StripeService) ConfirmPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentConfirmParams{}
	return paymentintent.Confirm(paymentIntentID, params)
}

// GetPaymentIntent récupère un PaymentIntent
func (s *StripeService) GetPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(paymentIntentID, nil)
}

// CreateTransfer crée un transfert vers un compte connecté
func (s *StripeService) CreateTransfer(amount int64, currency, destination string, metadata map[string]string) (*stripe.Transfer, error) {
	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(currency),
		Destination: stripe.String(destination),
		Metadata:    metadata,
	}

	return transfer.New(params)
}

// CreateSubscription crée un abonnement Stripe récurrent
func (s *StripeService) CreateSubscription(customerID, priceID string, metadata map[string]string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		Metadata: metadata,
	}

	return subscription.New(params)
}

// CancelSubscription annule un abonnement Stripe
func (s *StripeService) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{}
	return subscription.Cancel(subscriptionID, params)
}

// GetSubscription récupère un abonnement Stripe
func (s *StripeService) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return subscription.Get(subscriptionID, nil)
}

// CreatePrice crée un prix Stripe pour un plan d'abonnement
func (s *StripeService) CreatePrice(amount int64, currency, productID string, interval string) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		UnitAmount: stripe.Int64(amount),
		Currency:   stripe.String(currency),
		Product:    stripe.String(productID),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(interval), // "month" ou "year"
		},
	}

	return price.New(params)
}

// CreateProduct crée un produit Stripe
func (s *StripeService) CreateProduct(name, description string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Type:        stripe.String("service"),
	}

	return product.New(params)
}

// RefundPayment rembourse un paiement
func (s *StripeService) RefundPayment(paymentIntentID string, amount *int64) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	if amount != nil {
		params.Amount = stripe.Int64(*amount)
	}

	return refund.New(params)
}

// ListCustomerPaymentMethods liste toutes les méthodes de paiement d'un client
func (s *StripeService) ListCustomerPaymentMethods(customerID string, paymentMethodType string) ([]*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(paymentMethodType),
	}

	iter := paymentmethod.List(params)
	var methods []*stripe.PaymentMethod

	for iter.Next() {
		methods = append(methods, iter.PaymentMethod())
	}

	return methods, iter.Err()
}

// CreateConnectedAccount crée un compte connecté Stripe pour un créateur
func (s *StripeService) CreateConnectedAccount(email, country string) (*stripe.Account, error) {
	params := &stripe.AccountParams{
		Type:    stripe.String("express"),
		Email:   stripe.String(email),
		Country: stripe.String(country),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
	}

	return account.New(params)
}

// CreatePayoutToConnectedAccount crée un versement vers un compte connecté
func (s *StripeService) CreatePayoutToConnectedAccount(amount int64, currency, accountID string) (*stripe.Payout, error) {
	params := &stripe.PayoutParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
	}

	// Configurer le contexte pour le compte connecté
	params.SetStripeAccount(accountID)

	return payout.New(params)
}

// ConstructWebhookEvent construit et vérifie un événement webhook
func (s *StripeService) ConstructWebhookEvent(payload []byte, sig, endpointSecret string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sig, endpointSecret)
}

// FormatAmountForStripe convertit un montant en centimes pour Stripe
func (s *StripeService) FormatAmountForStripe(amount float64) int64 {
	return int64(amount * 100)
}

// FormatAmountFromStripe convertit un montant de centimes depuis Stripe
func (s *StripeService) FormatAmountFromStripe(amount int64) float64 {
	return float64(amount) / 100
}

// CalculateApplicationFee calcule les frais d'application (par exemple 10%)
func (s *StripeService) CalculateApplicationFee(amount float64, feePercentage float64) int64 {
	fee := amount * feePercentage / 100
	return s.FormatAmountForStripe(fee)
}

// PaymentMethodToStruct convertit une PaymentMethod Stripe en structure utilisable
type PaymentMethodInfo struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	ExpiryMonth int64  `json:"expiry_month"`
	ExpiryYear  int64  `json:"expiry_year"`
}

func (s *StripeService) PaymentMethodToStruct(pm *stripe.PaymentMethod) PaymentMethodInfo {
	info := PaymentMethodInfo{
		ID:   pm.ID,
		Type: string(pm.Type),
	}

	if pm.Card != nil {
		info.Last4 = pm.Card.Last4
		info.Brand = string(pm.Card.Brand)
		info.ExpiryMonth = pm.Card.ExpMonth
		info.ExpiryYear = pm.Card.ExpYear
	}

	return info
}
