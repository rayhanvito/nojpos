import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/mock/mock_seed_data.dart';
import '../models/product.dart';

final productRepositoryProvider = Provider<ProductRepository>(
  (ref) => const MockProductRepository(),
);

abstract interface class ProductRepository {
  List<String> getCategories();

  List<Product> getProducts();

  Product createProduct({
    required String name,
    required String category,
    required int price,
  });
}

class MockProductRepository implements ProductRepository {
  const MockProductRepository();

  @override
  List<String> getCategories() {
    return MockSeedData.categories.map((category) => category.name).toList();
  }

  @override
  List<Product> getProducts() {
    return const [
      ...MockSeedData.products,
      Product(
        id: 'p3',
        name: 'Bakso Komplit',
        category: 'Makanan',
        price: 17000,
        imageUrl:
            'https://images.unsplash.com/photo-1621996346565-e3dbc646d9a9?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p4',
        name: 'Cokelat Es',
        category: 'Minuman',
        price: 10000,
        imageUrl:
            'https://images.unsplash.com/photo-1572490122747-3968b75cc699?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p5',
        name: 'Kentang Goreng',
        category: 'Makanan',
        price: 10000,
        imageUrl:
            'https://images.unsplash.com/photo-1573080496219-bb080dd4f877?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p6',
        name: 'Kopi Susu',
        category: 'Minuman',
        price: 12000,
        imageUrl:
            'https://images.unsplash.com/photo-1461023058943-07fcbe16d735?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p7',
        name: 'Mie Ayam',
        category: 'Makanan',
        price: 18000,
        imageUrl:
            'https://images.unsplash.com/photo-1612929633738-8fe44f7ec841?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p8',
        name: 'Mix Platter',
        category: 'Paket',
        price: 20000,
        badge: 'Paket',
        imageUrl:
            'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p9',
        name: 'Nasi Goreng',
        category: 'Makanan',
        price: 17000,
        isFavorite: true,
        imageUrl:
            'https://images.unsplash.com/photo-1512058564366-18510be2db19?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p10',
        name: 'Nasi Putih',
        category: 'Makanan',
        price: 5000,
        imageUrl:
            'https://images.unsplash.com/photo-1536304993881-ff6e9eefa2a6?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p11',
        name: 'Paket Bakso + Es Teh',
        category: 'Paket',
        price: 17000,
        badge: 'Promo',
        imageUrl:
            'https://images.unsplash.com/photo-1540189549336-e6e99c3679fe?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p12',
        name: 'Potong Rambut',
        category: 'Layanan',
        price: 35000,
        badge: 'Layanan',
        imageUrl:
            'https://images.unsplash.com/photo-1621605815971-fbc98d665033?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p13',
        name: 'Soto Ayam',
        category: 'Makanan',
        price: 15000,
        imageUrl:
            'https://images.unsplash.com/photo-1569718212165-3a8278d5f624?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p14',
        name: 'Teh Manis',
        category: 'Minuman',
        price: 8000,
        imageUrl:
            'https://images.unsplash.com/photo-1499638673689-79a0b5115d87?auto=format&fit=crop&w=500&q=80',
      ),
      Product(
        id: 'p15',
        name: 'Telur Dadar',
        category: 'Makanan',
        price: 5000,
        imageUrl:
            'https://images.unsplash.com/photo-1525351484163-7529414344d8?auto=format&fit=crop&w=500&q=80',
      ),
    ];
  }

  @override
  Product createProduct({
    required String name,
    required String category,
    required int price,
  }) {
    return Product(
      id: 'product-${DateTime.now().microsecondsSinceEpoch}',
      name: name,
      category: category,
      price: price,
      imageUrl: '',
    );
  }
}
