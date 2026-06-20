import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';
import '../models/product.dart';

final productRepositoryProvider = Provider<ProductRepository>(
  (ref) => ApiProductRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class ProductRepository {
  Future<ProductCatalog> getCatalog({String? search, String? category});

  Future<Product> createProduct({
    String? outletId,
    String? productCategoryId,
    required String name,
    String? category,
    String? barcode,
    required int price,
    bool trackStock = true,
  });

  Future<Product> updateProduct({
    required String id,
    String? outletId,
    String? productCategoryId,
    required String name,
    String? category,
    String? barcode,
    required int price,
    bool trackStock = true,
  });

  Future<ProductCategory> createCategory({required String name});

  Future<ProductCategory> updateCategory({
    required String id,
    required String name,
  });
}

class ProductCatalog {
  const ProductCatalog({required this.categoryItems, required this.products});

  final List<ProductCategory> categoryItems;
  final List<Product> products;

  List<String> get categories => [
    for (final category in categoryItems) category.name,
  ];
}

class ApiProductRepository implements ProductRepository {
  const ApiProductRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<ProductCatalog> getCatalog({String? search, String? category}) async {
    final categoryItems = await _getCategories();
    final productResponse = await _apiClient.get<Object?>(
      '/products',
      queryParameters: {
        if (search != null && search.trim().isNotEmpty) 'search': search.trim(),
        if (category != null &&
            category.trim().isNotEmpty &&
            !_isSystemCategoryName(category))
          'category': category.trim(),
      },
    );
    final categoryNamesById = {
      for (final category in categoryItems)
        if (category.id.isNotEmpty) category.id: category.name,
    };
    final products = [
      for (final item in _listFromPayload(productResponse.data, 'products'))
        _productFromJson(item, categoryNamesById),
    ];

    return ProductCatalog(
      categoryItems: _mergeCategoryItems(categoryItems, products),
      products: products,
    );
  }

  Future<List<ProductCategory>> _getCategories() async {
    final response = await _apiClient.get<Object?>('/categories');
    final categoryItems = <ProductCategory>[
      const ProductCategory(id: '__system_all', name: 'Semua'),
      const ProductCategory(id: '__system_favorite', name: 'Favorit'),
    ];
    for (final item in _listFromPayload(response.data, 'categories')) {
      final category = _categoryFromJson(item);
      if (category.name.isEmpty ||
          categoryItems.any((existing) => existing.name == category.name)) {
        continue;
      }
      categoryItems.add(category);
    }
    return categoryItems;
  }

  @override
  Future<Product> createProduct({
    String? outletId,
    String? productCategoryId,
    required String name,
    String? category,
    String? barcode,
    required int price,
    bool trackStock = true,
  }) async {
    final trimmedBarcode = barcode?.trim();
    final categoryNamesById = productCategoryId == null
        ? const <String, String>{}
        : <String, String>{productCategoryId: category ?? productCategoryId};
    final data = <String, Object?>{
      'outlet_id': outletId,
      'product_category_id': productCategoryId,
      'name': name,
      'price': price,
      'track_stock': trackStock,
    };
    if (trimmedBarcode != null && trimmedBarcode.isNotEmpty) {
      data['barcode'] = trimmedBarcode;
    }
    final response = await _apiClient.post<Map<String, Object?>>(
      '/products',
      data: data,
    );
    return _productFromJson(response.data, categoryNamesById);
  }

  @override
  Future<Product> updateProduct({
    required String id,
    String? outletId,
    String? productCategoryId,
    required String name,
    String? category,
    String? barcode,
    required int price,
    bool trackStock = true,
  }) async {
    final trimmedBarcode = barcode?.trim();
    final categoryNamesById = productCategoryId == null
        ? const <String, String>{}
        : <String, String>{productCategoryId: category ?? productCategoryId};
    final data = <String, Object?>{
      'outlet_id': outletId,
      'product_category_id': productCategoryId,
      'name': name,
      'price': price,
      'track_stock': trackStock,
    };
    if (trimmedBarcode != null && trimmedBarcode.isNotEmpty) {
      data['barcode'] = trimmedBarcode;
    }
    final response = await _apiClient.put<Map<String, Object?>>(
      '/products/$id',
      data: data,
    );
    return _productFromJson(response.data, categoryNamesById);
  }

  @override
  Future<ProductCategory> createCategory({required String name}) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/categories',
      data: {'name': name},
    );
    return _categoryFromJson(response.data);
  }

  @override
  Future<ProductCategory> updateCategory({
    required String id,
    required String name,
  }) async {
    final response = await _apiClient.put<Map<String, Object?>>(
      '/categories/$id',
      data: {'name': name},
    );
    return _categoryFromJson(response.data);
  }
}

ProductCategory _categoryFromJson(Object? value) {
  final json = _asMap(value);
  return ProductCategory(
    id: (json['id'] as String?) ?? '',
    name: (json['name'] as String?) ?? '',
  );
}

Product _productFromJson(Object? value, Map<String, String> categoryNamesById) {
  final json = _asMap(value);
  final categoryId =
      (json['product_category_id'] ?? json['category_id']) as String?;
  final category = categoryId == null
      ? (json['category'] as String?) ?? 'Lainnya'
      : categoryNamesById[categoryId] ??
            (json['category'] as String?) ??
            'Lainnya';
  return Product(
    id: (json['id'] as String?) ?? '',
    name: (json['name'] as String?) ?? 'Produk',
    category: category,
    categoryId: categoryId,
    price: (json['price'] as num?)?.toInt() ?? 0,
    imageUrl: '',
    badge: (json['barcode'] as String?)?.isEmpty ?? true
        ? null
        : json['barcode'] as String?,
  );
}

List<ProductCategory> _mergeCategoryItems(
  List<ProductCategory> categoryItems,
  List<Product> products,
) {
  final merged = <ProductCategory>[];
  for (final category in categoryItems) {
    if (category.name.isEmpty ||
        merged.any((existing) => existing.name == category.name)) {
      continue;
    }
    merged.add(category);
  }
  for (final product in products) {
    if (product.category.isEmpty ||
        merged.any((category) => category.name == product.category)) {
      continue;
    }
    merged.add(
      ProductCategory(id: product.categoryId ?? '', name: product.category),
    );
  }
  return merged;
}

List<Object?> _listFromPayload(Object? payload, String key) {
  if (payload is List) return List<Object?>.from(payload);
  final json = _asMap(payload);
  final value = json[key] ?? json['items'];
  if (value is List) return List<Object?>.from(value);
  return const [];
}

bool _isSystemCategoryName(String category) {
  return category == 'Semua' || category == 'Favorit';
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
