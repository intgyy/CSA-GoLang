package router

import (
	"client/api"
	"client/middleware"
	"github.com/gin-gonic/gin"
)

func InitUserRouter(Router *gin.RouterGroup) {

	Router.POST("register", api.UserRegister)
	Router.POST("login", api.UserLogin)
	UserRouter := Router.Group("user", middleware.JWTAuthMiddleware())
	{
		UserRouter.GET("search", api.SearchGoodsList)
		UserRouter.GET("favorite", api.FavoriteGoods)
		UserRouter.GET("searchFavorite", api.GetFavoriteGoods)
		UserRouter.GET("clearCar", api.ClearGoodsInCar)
		UserRouter.GET("buyInCar", api.BuyAllGoodsInCar)
		UserRouter.GET("addInCar", api.AddGoodsInCar)
		UserRouter.GET("reduceInCar", api.DeleteGoodsInCar)
		UserRouter.GET("buy", api.BuyGoods)
		UserRouter.POST("publish", api.PublishGoods)
		UserRouter.GET("delete", api.DeleteGoods)
		UserRouter.POST("update", api.UpdateGoods)
	}

}
