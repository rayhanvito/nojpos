import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/nojpos_models.dart';
import '../models/cart_item.dart';
import '../models/product.dart';
import '../repositories/product_repository.dart';

final categoriesProvider = Provider<List<String>>((ref) {
  return ref.watch(posCatalogProvider).categories;
});

final categoryItemsProvider = Provider<List<ProductCategory>>((ref) {
  return ref.watch(posCatalogProvider).categoryItems;
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
  const PosCatalogState({
    required this.categoryItems,
    required this.products,
    this.isLoading = false,
    this.errorMessage,
  });

  final List<ProductCategory> categoryItems;
  final List<Product> products;
  final bool isLoading;
  final String? errorMessage;

  List<String> get categories => [
    for (final category in categoryItems) category.name,
  ];

  PosCatalogState copyWith({
    List<ProductCategory>? categoryItems,
    List<Product>? products,
    bool? isLoading,
    String? errorMessage,
    bool clearError = false,
  }) {
    return PosCatalogState(
      categoryItems: categoryItems ?? this.categoryItems,
      products: products ?? this.products,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }
}

class PosCatalogNotifier extends Notifier<PosCatalogState> {
  @override
  PosCatalogState build() {
    Future<void>.microtask(load);
    return const PosCatalogState(
      categoryItems: [
        ProductCategory(id: '__system_all', name: 'Semua'),
        ProductCategory(id: '__system_favorite', name: 'Favorit'),
      ],
      products: [],
      isLoading: true,
    );
  }

  Future<void> load({String? search, String? category}) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final catalog = await ref
          .read(productRepositoryProvider)
          .getCatalog(search: search, category: category);
      state = state.copyWith(
        categoryItems: catalog.categoryItems,
        products: catalog.products,
        isLoading: false,
      );
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: _messageFor(error),
      );
    }
  }

  Future<Product> addProduct({
    required String outletId,
    required String name,
    required ProductCategory category,
    required int price,
  }) async {
    final product = ref
        .read(productRepositoryProvider)
        .createProduct(
          outletId: outletId,
          productCategoryId: _categoryIdForWrite(category),
          name: name,
          category: category.name,
          price: price,
        );
    final created = await product;
    final categoryItems =
        state.categoryItems.any(
          (item) => item.id == category.id || item.name == category.name,
        )
        ? state.categoryItems
        : [...state.categoryItems, category];
    state = state.copyWith(
      categoryItems: categoryItems,
      products: [created, ...state.products],
    );
    return created;
  }

  Future<Product> updateProduct({
    required Product product,
    required String outletId,
    required String name,
    required ProductCategory category,
    required int price,
  }) async {
    final updated = await ref
        .read(productRepositoryProvider)
        .updateProduct(
          id: product.id,
          outletId: outletId,
          productCategoryId: _categoryIdForWrite(category),
          name: name,
          category: category.name,
          price: price,
        );
    final normalized = _productWithCategoryFallback(updated, category);
    state = state.copyWith(
      products: [
        for (final existing in state.products)
          existing.id == product.id ? normalized : existing,
      ],
    );
    return normalized;
  }

  Future<ProductCategory> addCategory(String name) async {
    final normalized = name.trim();
    if (normalized.isEmpty) {
      return const ProductCategory(id: '', name: '');
    }
    final category = await ref
        .read(productRepositoryProvider)
        .createCategory(name: normalized);
    state = state.copyWith(categoryItems: [...state.categoryItems, category]);
    return category;
  }

  Future<ProductCategory> updateCategory({
    required String id,
    required String oldName,
    required String name,
  }) async {
    final updated = await ref
        .read(productRepositoryProvider)
        .updateCategory(id: id, name: name);
    state = state.copyWith(
      categoryItems: [
        for (final category in state.categoryItems)
          category.id == id || category.name == oldName ? updated : category,
      ],
      products: [
        for (final product in state.products)
          product.categoryId == id || product.category == oldName
              ? Product(
                  id: product.id,
                  name: product.name,
                  category: updated.name,
                  categoryId: updated.id,
                  price: product.price,
                  imageUrl: product.imageUrl,
                  badge: product.badge,
                  isFavorite: product.isFavorite,
                )
              : product,
      ],
    );
    return updated;
  }
}

String? _categoryIdForWrite(ProductCategory category) {
  if (category.id.startsWith('__system_') || category.id.isEmpty) return null;
  return category.id;
}

Product _productWithCategoryFallback(
  Product product,
  ProductCategory category,
) {
  if (product.categoryId != null && product.category == category.name) {
    return product;
  }
  return Product(
    id: product.id,
    name: product.name,
    category: product.category == 'Lainnya' ? category.name : product.category,
    categoryId: product.categoryId ?? _categoryIdForWrite(category),
    price: product.price,
    imageUrl: product.imageUrl,
    badge: product.badge,
    isFavorite: product.isFavorite,
  );
}

String _messageFor(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
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

  void setDiscount(Product product, int discount) {
    state = [
      for (final item in state)
        if (item.product.id == product.id)
          item.copyWith(discount: discount.clamp(0, item.subtotal))
        else
          item,
    ];
  }

  void replaceWith(List<CartItem> items) {
    state = items;
  }

  void clear() {
    state = [];
  }
}

final cartTotalProvider = Provider<int>((ref) {
  final subtotal = ref
      .watch(cartProvider)
      .fold(0, (sum, item) => sum + item.total);
  final cartDiscount = ref.watch(checkoutDetailsProvider).cartDiscount;
  return (subtotal - cartDiscount).clamp(0, subtotal);
});

final cartSubtotalProvider = Provider<int>((ref) {
  return ref.watch(cartProvider).fold(0, (sum, item) => sum + item.subtotal);
});

final itemDiscountTotalProvider = Provider<int>((ref) {
  return ref.watch(cartProvider).fold(0, (sum, item) => sum + item.discount);
});

final checkoutDetailsProvider =
    NotifierProvider<CheckoutDetailsNotifier, CheckoutDetailsState>(
      CheckoutDetailsNotifier.new,
    );

class CheckoutDetailsState {
  const CheckoutDetailsState({
    this.cartDiscount = 0,
    this.notes = '',
    this.servedBy,
  });

  final int cartDiscount;
  final String notes;
  final Employee? servedBy;

  CheckoutDetailsState copyWith({
    int? cartDiscount,
    String? notes,
    Employee? servedBy,
    bool clearServedBy = false,
  }) {
    return CheckoutDetailsState(
      cartDiscount: cartDiscount ?? this.cartDiscount,
      notes: notes ?? this.notes,
      servedBy: clearServedBy ? null : servedBy ?? this.servedBy,
    );
  }
}

class CheckoutDetailsNotifier extends Notifier<CheckoutDetailsState> {
  @override
  CheckoutDetailsState build() => const CheckoutDetailsState();

  void setCartDiscount(int discount, int subtotalAfterItemDiscount) {
    state = state.copyWith(
      cartDiscount: discount.clamp(0, subtotalAfterItemDiscount),
    );
  }

  void setNotes(String notes) {
    state = state.copyWith(notes: notes.trim());
  }

  void setServedBy(Employee? employee) {
    state = state.copyWith(servedBy: employee, clearServedBy: employee == null);
  }

  void clear() {
    state = const CheckoutDetailsState();
  }
}
