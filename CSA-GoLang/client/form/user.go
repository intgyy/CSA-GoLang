package form

type RegisterForm struct {
	UserName string  `form:"username" json:"userName" binding:"required"`
	PassWord string  `form:"password" json:"passWord" binding:"required,min=8,max=15"`
	Price    float32 `form:"price" json:"price" binding:"required"`
}
type PassWordLoginForm struct {
	UserName string `form:"username" json:"userName" binding:"required"`
	PassWord string `form:"password" json:"passWord" binding:"required,min=8,max=15"`
}
type GoodsForm struct {
	Title    string  `form:"title" json:"title" binding:"required"`
	Describe string  `form:"describe" json:"describe" binding:"required"`
	Price    float32 `form:"price" json:"price" binding:"required"`
}
type UpdateGoodsForm struct {
	Title    string  `form:"title" json:"title" binding:"required"`
	Describe string  `form:"describe" json:"describe" binding:"required"`
	Price    float32 `form:"price" json:"price" binding:"required"`
	GoodsID  int     `form:"goodsId" json:"goodsId" binding:"required"`
}
