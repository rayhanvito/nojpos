class Product {
  const Product({
    required this.id,
    required this.name,
    required this.category,
    required this.price,
    required this.imageUrl,
    this.badge,
    this.isFavorite = false,
  });

  final String id;
  final String name;
  final String category;
  final int price;
  final String imageUrl;
  final String? badge;
  final bool isFavorite;
}
