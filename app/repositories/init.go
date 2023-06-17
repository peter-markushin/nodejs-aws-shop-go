package repositories

func Init() {
	ProductRepositoryImpl = new(ProductRepository)
	StockReposityryImpl = new(StockRepository)
}
