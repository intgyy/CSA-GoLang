package handler

import (
	"CSA-GoLang/server/global"
	"CSA-GoLang/server/model"
	"CSA-GoLang/server/proto"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type UserServer struct {
	proto.UnimplementedUserServer
}

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))
}

func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.UserInfo, error) {
	var user model.User

	result := global.DB.Where(&model.User{Username: req.Username}).First(&user)

	if result.RowsAffected != 0 {
		zap.S().Errorw("创建用户失败")
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}
	user.Password = genMd5(req.Password)
	user.Username = req.Username
	user.Balance = float64(req.Price)
	user.Role = 0
	result = global.DB.Create(&user)
	if result.Error != nil {
		zap.S().Errorw("创建用户失败")
		return nil, status.Errorf(codes.Internal, "创建用户失败")
	}
	resp := proto.UserInfo{
		Id:       int32(user.ID),
		Username: user.Username,
		Password: user.Password,
		Price:    float32(user.Balance),
	}
	return &resp, nil

}
func (s *UserServer) GetGoods(ctx context.Context, req *proto.GetGoodsRequest) (*proto.GoodsList, error) {
	var goodsList []model.Goods
	result := global.DB.Where(&model.Goods{Title: req.Title}).Find(&goodsList)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	resp := &proto.GoodsList{}
	for _, goods := range goodsList {
		goodsreq := proto.GoodsInfo{
			Id:      int32(goods.ID),
			Title:   goods.Title,
			Price:   float32(goods.Price),
			Content: goods.Description,
		}
		resp.GoodsList = append(resp.GoodsList, &goodsreq)
	}
	return resp, nil

}
func (s *UserServer) FavoriteGoods(ctx context.Context, req *proto.GoodsUserRequest) (*proto.SuccessResponse, error) {
	var goods model.Goods
	var user model.User
	var goodsList []model.Goods
	err := global.DB.Model(&user).Association("Favorites").Find(&goodsList)
	if err != nil {
		fmt.Println(err)
		zap.S().Errorw("收藏商品失败", err)
		return nil, status.Errorf(codes.Internal, "收藏商品失败")
	}
	for _, goods := range goodsList {
		if int32(goods.ID) == req.GoodsId {
			zap.S().Errorw("已经收藏过了")
			return nil, status.Errorf(codes.AlreadyExists, "已经收藏过了")
		}
	}
	result := global.DB.First(&goods, req.GoodsId)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	err = global.DB.Model(&user).Association("Favorites").Append(&goods)

	if err != nil {
		fmt.Println(err)
		zap.S().Errorw("收藏商品失败")
		return nil, status.Errorf(codes.Internal, "收藏商品失败")
	}
	resp := proto.SuccessResponse{
		Success: true,
	}
	return &resp, nil
}
func (s *UserServer) GetFavoriteGoods(ctx context.Context, req *proto.NameRequest) (*proto.GoodsList, error) {
	var user model.User
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	var goodsList []model.Goods
	err := global.DB.Model(&user).Association("Favorites").Find(&goodsList)
	if err != nil {
		zap.S().Errorw("没有物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	resp := &proto.GoodsList{}
	for _, goods := range goodsList {
		goodsreq := proto.GoodsInfo{
			Id:      int32(goods.ID),
			Title:   goods.Title,
			Price:   float32(goods.Price),
			Content: goods.Description,
		}
		resp.GoodsList = append(resp.GoodsList, &goodsreq)
	}
	return resp, nil

}
func (s *UserServer) ClearGoodsInCar(ctx context.Context, req *proto.NameRequest) (*proto.SuccessResponse, error) {
	var user model.User
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	result := global.DB.Unscoped().Where("user_id = ?", user.ID).Delete(&model.Cart{})

	if result.Error != nil {
		zap.S().Errorw("清空商品失败")
		return nil, status.Errorf(codes.Internal, "清空购物车失败")
	}
	resp := proto.SuccessResponse{Success: true}
	return &resp, nil
}
func (s *UserServer) BuyAllGoodsInCar(ctx context.Context, req *proto.NameRequest) (*proto.SuccessResponse, error) {
	var user model.User
	var carts []model.Cart
	var goods model.Goods
	var allPrice float64 = 0
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	global.DB.Where("user_id = ?", user.ID).Find(&carts)
	for _, cart := range carts {
		global.DB.First(&goods, cart.GoodsID)
		price := float64(cart.Quantity) * goods.Price
		allPrice = allPrice + price
	}

	if user.Balance < allPrice {
		zap.S().Errorw("余额不足")
		return nil, status.Errorf(codes.Internal, "余额不足")
	}
	result := global.DB.Unscoped().Where("user_id = ?", user.ID).Delete(&model.Cart{})
	if result.Error != nil {
		zap.S().Errorw("清空商品失败")
		return nil, status.Errorf(codes.Internal, "清空购物车失败")
	}
	user.Balance = user.Balance - allPrice
	global.DB.Save(&user)
	resp := proto.SuccessResponse{Success: true}
	return &resp, nil
}
func (s *UserServer) AddGoodsInCar(ctx context.Context, req *proto.GoodsUserRequest) (*proto.CartInfo, error) {
	var goods model.Goods
	var user model.User
	result := global.DB.First(&goods, req.GoodsId)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	global.DB.Where(&model.User{Username: req.Name}).First(&user)

	var cart model.Cart
	result = global.DB.Where(&model.Cart{UserID: user.ID, GoodsID: goods.ID}).First(&cart)
	if result.RowsAffected == 0 {
		cart.Quantity = 1
	} else {
		cart.Quantity = cart.Quantity + 1
	}
	cart.GoodsID = goods.ID
	cart.UserID = user.ID
	global.DB.Save(&cart)

	resp := proto.CartInfo{
		UserId:   int32(user.ID),
		GoodsId:  int32(goods.ID),
		Quantity: int32(cart.Quantity),
	}
	return &resp, nil
}
func (s *UserServer) ReduceGoodsInCar(ctx context.Context, req *proto.GoodsUserRequest) (*proto.CartInfo, error) {
	var goods model.Goods
	var user model.User
	result := global.DB.First(&goods, req.GoodsId)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	global.DB.Where(&model.User{Username: req.Name}).First(&user)

	var cart model.Cart
	result = global.DB.Where(&model.Cart{UserID: user.ID, GoodsID: goods.ID}).First(&cart)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}
	if cart.Quantity < 1 {
		global.DB.Unscoped().Delete(&cart)
		return nil, status.Errorf(codes.Internal, "已经全部删完了")
	}
	if cart.Quantity == 1 {
		global.DB.Unscoped().Delete(&cart)
		resp := proto.CartInfo{Quantity: 0}
		return &resp, nil
	}
	if cart.Quantity > 1 {
		cart.Quantity = cart.Quantity - 1
	}

	global.DB.Save(&cart)

	resp := proto.CartInfo{
		UserId:   int32(user.ID),
		GoodsId:  int32(goods.ID),
		Quantity: int32(cart.Quantity),
	}
	return &resp, nil
}
func (s *UserServer) BuyGoods(ctx context.Context, req *proto.GoodsUserRequest) (*proto.SuccessResponse, error) {
	var goods model.Goods
	var user model.User
	result := global.DB.First(&goods, req.GoodsId)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	if user.Balance < goods.Price {
		zap.S().Errorw("余额不足")
		return nil, status.Errorf(codes.Internal, "余额不足")
	}
	user.Balance = user.Balance - goods.Price
	result = global.DB.Save(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "购买失败")
	}
	resp := proto.SuccessResponse{
		Success: true,
	}
	return &resp, nil
}
func (s *UserServer) PublishGoods(ctx context.Context, req *proto.CreateGoodsRequest) (*proto.GoodsInfo, error) {
	var goods model.Goods
	var user model.User
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	if user.Role != 1 {
		return nil, status.Errorf(codes.Internal, "权限不够")
	}
	goods.Title = req.Title
	goods.Price = float64(req.Price)
	goods.Description = req.Content
	goods.UserID = user.ID
	global.DB.Create(&goods)
	resp := proto.GoodsInfo{
		Id:      int32(goods.ID),
		Title:   goods.Title,
		Price:   float32(goods.Price),
		Content: goods.Description,
	}
	return &resp, nil
}
func (s *UserServer) DeleteGoods(ctx context.Context, req *proto.GoodsUserRequest) (*proto.SuccessResponse, error) {
	var goods model.Goods
	var user model.User
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	if user.Role != 1 {
		return nil, status.Errorf(codes.Internal, "权限不够")
	}
	result := global.DB.First(&goods, req.GoodsId)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	if user.ID != goods.UserID {
		zap.S().Errorw("没有权限")
		return nil, status.Errorf(codes.NotFound, "没有权限")
	}
	goods_id := goods.ID
	result = global.DB.Delete(&goods)
	if result.Error != nil {
		zap.S().Errorw("删除商品失败")
		return nil, status.Errorf(codes.NotFound, "删除商品失败")
	}
	global.DB.Unscoped().Where("goods_id = ?", goods_id).Delete(&model.Cart{})
	resp := proto.SuccessResponse{
		Success: true,
	}
	return &resp, nil
}
func (s *UserServer) UpdateGoods(ctx context.Context, req *proto.UpdateGoodsRequest) (*proto.GoodsInfo, error) {
	var goods model.Goods
	var user model.User
	global.DB.Where(&model.User{Username: req.Name}).First(&user)
	if user.Role != 1 {
		return nil, status.Errorf(codes.Internal, "权限不够")
	}
	result := global.DB.First(&goods, req.GoodsId)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到物品")
		return nil, status.Errorf(codes.NotFound, "物品不存在")
	}
	if user.ID != goods.UserID {
		zap.S().Errorw("没有权限")
		return nil, status.Errorf(codes.NotFound, "没有权限")
	}
	goods.Title = req.Title
	goods.Price = float64(req.Price)
	goods.Description = req.Content
	result = global.DB.Updates(&goods)
	if result.Error != nil {
		zap.S().Errorw("更新商品失败")
		return nil, status.Errorf(codes.Internal, "更新商品失败")
	}
	resp := proto.GoodsInfo{
		Id:      int32(goods.ID),
		Title:   goods.Title,
		Price:   float32(goods.Price),
		Content: goods.Description,
	}
	return &resp, nil
}
func (s *UserServer) CheckPassword(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.SuccessResponse, error) {
	password := genMd5(req.Password)
	resp := proto.SuccessResponse{Success: false}
	if password == req.EncryptedPassword {
		resp.Success = true
		return &resp, nil
	}
	return &resp, nil
}
func (s *UserServer) GetUserByName(ctx context.Context, req *proto.NameRequest) (*proto.UserInfo, error) {
	var user model.User
	result := global.DB.Where(&model.User{Username: req.Name}).First(&user)
	if result.RowsAffected == 0 {
		zap.S().Errorw("没有搜到用户")
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	resp := proto.UserInfo{
		Id:       int32(user.ID),
		Username: user.Username,
		Password: user.Password,
		Price:    float32(user.Balance),
	}
	return &resp, nil
}
