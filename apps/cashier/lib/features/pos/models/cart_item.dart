import 'product.dart';

class CartItem {
  const CartItem({
    required this.product,
    required this.quantity,
    this.discount = 0,
  });

  final Product product;
  final int quantity;
  final int discount;

  int get subtotal => product.price * quantity;

  int get total => (subtotal - discount).clamp(0, subtotal);

  CartItem copyWith({int? quantity, int? discount}) {
    return CartItem(
      product: product,
      quantity: quantity ?? this.quantity,
      discount: discount ?? this.discount,
    );
  }
}
