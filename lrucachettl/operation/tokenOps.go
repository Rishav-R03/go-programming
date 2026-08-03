package operation

import "lrucachettl/model"

func NewTokenCache() *model.TokenCache {
	return &model.TokenCache{}
}

func GetToken(tc *model.TokenCache, sessionString string) {
	tc.Mu.RLock()

}
