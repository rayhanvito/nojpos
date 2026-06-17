import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/cart_item.dart';
import '../models/product.dart';
import '../repositories/product_repository.dart';

final categoriesProvider = Provider<List<String>>((ref) {
  return ref.watch(posCatalogProvider).categories;
});

final selectedCategoryProvider =
    NotifierProvider<SelectedCategoryNotifier, String>(
      SelectedCategoryNotifier.new,
    );

class SelectedCategoryNotifier extends Notifier<String> {
  @override
  String build() => 'Semua';

  void select(String category) {
    state = category;
  }
}

final posCatalogProvider =
    NotifierProvider<PosCatalogNotifier, PosCatalogState>(
      PosCatalogNotifier.new,
    );

class PosCatalogState {
  const PosCatalogState({required this.categories, required this.products});

  final List<String> categories;
  final List<Product> products;

  PosCatalogState copyWith({
    List<String>? categories,
    List<Product>? products,
  }) {
    return PosCatalogState(
      categories: categories ?? this.categories,
      products: products ?? this.products,
    );
  }
}

class PosCatalogNotifier extends Notifier<PosCatalogState> {
  @override
  PosCatalogState build() {
    final productRepository = ref.read(productRepositoryProvider);
    return PosCatalogState(
      categories: productRepository.getCategories(),
      products: productRepository.getProducts(),
    );
  }

  Product addProduct({
    required String name,
    required String category,
    required int price,
  }) {
    final product = ref
        .read(productRepositoryProvider)
        .createProduct(name: name, category: category, price: price);
    final categories = state.categories.contains(category)
        ? state.categories
        : [...state.categories, category];
    state = state.copyWith(
      categories: categories,
      products: [product, ...state.products],
    );
    return product;
  }

  void addCategory(String name) {
    final normalized = name.trim();
    if (normalized.isEmpty || state.categories.contains(normalized)) return;
    state = state.copyWith(categories: [...state.categories, normalized]);
  }
}

final productsProvider = Provider<List<Product>>((ref) {
  return ref.watch(posCatalogProvider).products;
});

final productSearchQueryProvider =
    NotifierProvider<ProductSearchQueryNotifier, String>(
      ProductSearchQueryNotifier.new,
    );

class ProductSearchQueryNotifier extends Notifier<String> {
  @override
  String build() => '';

  void update(String query) {
    state = query;
  }
}

final filteredProductsProvider = Provider<List<Product>>((ref) {
  final category = ref.watch(selectedCategoryProvider);
  final query = ref.watch(productSearchQueryProvider).trim().toLowerCase();
  final products = ref.watch(productsProvider);
  final categorized = switch (category) {
    'Semua' => products,
    'Favorit' => products.where((product) => product.isFavorite).toList(),
    _ => products.where((product) => product.category == category).toList(),
  };
  if (query.isEmpty) {
    return categorized;
  }
  return categorized
      .where(
        (product) =>
            product.name.toLowerCase().contains(query) ||
            product.category.toLowerCase().contains(query) ||
            product.id.toLowerCase().contains(query),
      )
      .toList();
});

final cartProvider = NotifierProvider<CartNotifier, List<CartItem>>(
  CartNotifier.new,
);

class CartNotifier extends Notifier<List<CartItem>> {
  @override
  List<CartItem> build() => [];

  void add(Product product) {
    final index = state.indexWhere((item) => item.product.id == product.id);
    if (index == -1) {
      state = [...state, CartItem(product: product, quantity: 1)];
      return;
    }

    state = [
      for (final (itemIndex, item) in state.indexed)
        itemIndex == index ? item.copyWith(quantity: item.quantity + 1) : item,
    ];
  }

  void decrease(Product product) {
    state = [
      for (final item in state)
        if (item.product.id == product.id && item.quantity > 1)
          item.copyWith(quantity: item.quantity - 1)
        else if (item.product.id != product.id)
          item,
    ];
  }

  void remove(Product product) {
    state = state.where((item) => item.product.id != product.id).toList();
  }

  void replaceWith(List<CartItem> items) {
    state = items;
  }

  void clear() {
    state = [];
  }
}

final cartTotalProvider = Provider<int>((ref) {
  return ref.watch(cartProvider).fold(0, (sum, item) => sum + item.subtotal);
});
