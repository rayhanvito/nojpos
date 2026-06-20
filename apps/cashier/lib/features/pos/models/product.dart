class Product {
  const Product({
    required this.id,
    required this.name,
    required this.category,
    required this.price,
    required this.imageUrl,
    this.categoryId,
    this.badge,
    this.isFavorite = false,
  });

  final String id;
  final String name;
  final String category;
  final int price;
  final String imageUrl;
  final String? categoryId;
  final String? badge;
  final bool isFavorite;
}
