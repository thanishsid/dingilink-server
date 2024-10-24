package ctxt

import (
	"context"

	"github.com/thanishsid/dingilink-server/internal/db"
)

type ContextKey string

const (
	USER_INFO_CTX_KEY   ContextKey = "user.info"
	DB_CTX_KEY          ContextKey = "db"
	API_VERSION_CTX_KEY ContextKey = "api.version"
	DATALOADER_CTX_KEY  ContextKey = "dataloader"
)

func DBFromContext(ctx context.Context) db.DBQ {
	d, ok := ctx.Value(DB_CTX_KEY).(db.DBQ)
	if !ok {
		panic("db not found in ctx")
	}

	return d
}

// func StripeFromContext(ctx context.Context) *stripeclient.API {
// 	c, ok := ctx.Value(STRIPE_CTX_KEY).(*stripeclient.API)
// 	if !ok {
// 		panic("stripe client not found in ctx")
// 	}

// 	return c
// }

// func WalletFromContext(ctx context.Context) *commonmodel.Wallet {
// 	wallet, ok := ctx.Value(WALLET_CTX_KEY).(*commonmodel.Wallet)
// 	if ok {
// 		return wallet
// 	}

// 	return nil
// }

// func CommonDataLoaderFromContext(ctx context.Context) dloader.CommonDataLoader {
// 	d, ok := ctx.Value(COMMON_DATALOADER_CTX_KEY).(dloader.CommonDataLoader)
// 	if !ok {
// 		panic("common dataloader not found in ctx")
// 	}

// 	return d
// }

// func FoodDataLoaderFromContext(ctx context.Context) dloader.FoodDataLoader {
// 	d, ok := ctx.Value(FOOD_DATALOADER_CTX_KEY).(dloader.FoodDataLoader)
// 	if !ok {
// 		panic("food dataloader not found in ctx")
// 	}

// 	return d
// }
