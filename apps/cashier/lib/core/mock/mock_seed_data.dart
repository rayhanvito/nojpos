import '../../features/pos/models/product.dart';
import '../../shared/models/nojpos_models.dart';

class MockSeedData {
  const MockSeedData._();

  static final outlet = Outlet(
    id: 'outlet-1',
    name: 'Kedai Nusantara',
    isOnline: true,
  );

  static const employees = [
    Employee(id: 'employee-1', name: 'Rayhan Vito', role: 'Kasir', pin: '1234'),
    Employee(
      id: 'employee-2',
      name: 'Rayhan Vito Gustiansyah',
      role: 'Supervisor',
      pin: '1234',
    ),
  ];

  static const customers = [
    Customer(id: 'customer-1', name: 'Tanpa Pelanggan', group: 'Tanpa Grup'),
    Customer(
      id: 'customer-2',
      name: 'Adit Pratama',
      phone: '081234567890',
      group: 'Regular',
    ),
  ];

  static const categories = [
    ProductCategory(id: 'category-all', name: 'Semua'),
    ProductCategory(id: 'category-favorite', name: 'Favorit'),
    ProductCategory(id: 'category-food', name: 'Makanan'),
    ProductCategory(id: 'category-drink', name: 'Minuman'),
    ProductCategory(id: 'category-package', name: 'Paket'),
    ProductCategory(id: 'category-service', name: 'Layanan'),
  ];

  static const products = [
    Product(
      id: 'p1',
      name: 'Air Mineral Prima',
      category: 'Minuman',
      price: 3000,
      badge: 'Grosir',
      isFavorite: true,
      imageUrl:
          'https://images.unsplash.com/photo-1616118132534-381148898bb4?auto=format&fit=crop&w=500&q=80',
    ),
    Product(
      id: 'p2',
      name: 'Ayam Teriyaki',
      category: 'Makanan',
      price: 23000,
      isFavorite: true,
      imageUrl:
          'https://images.unsplash.com/photo-1603133872878-684f208fb84b?auto=format&fit=crop&w=500&q=80',
    ),
    Product(
      id: 'p8',
      name: 'Mix Platter',
      category: 'Paket',
      price: 20000,
      badge: 'Paket',
      imageUrl:
          'https://images.unsplash.com/photo-1543353071-087092ec393a?auto=format&fit=crop&w=500&q=80',
    ),
  ];
}
