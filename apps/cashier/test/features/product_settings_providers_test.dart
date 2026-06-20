import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/features/pos/models/product.dart';
import 'package:nojpos_tablet_ui/features/pos/providers/pos_providers.dart';
import 'package:nojpos_tablet_ui/features/pos/repositories/product_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  group('PosCatalogNotifier', () {
    late _FakeProductRepository repository;
    late ProviderContainer container;

    setUp(() {
      repository = _FakeProductRepository();
      container = ProviderContainer(
        overrides: [
          productRepositoryProvider.overrideWith((ref) => repository),
        ],
      );
    });

    tearDown(() {
      container.dispose();
    });

    test('should load categories from repository API seam', () async {
      await container.read(posCatalogProvider.notifier).load();

      expect(repository.loadCalls, greaterThan(0));
      expect(container.read(posCatalogProvider).categories, [
        'Semua',
        'Favorit',
        'Minuman',
        'Snack',
      ]);
    });

    test(
      'should send category id when creating and updating products',
      () async {
        await container.read(posCatalogProvider.notifier).load();
        final categories = container.read(posCatalogProvider).categoryItems;
        final drink = categories.singleWhere((item) => item.name == 'Minuman');
        final snack = categories.singleWhere((item) => item.name == 'Snack');

        final created = await container
            .read(posCatalogProvider.notifier)
            .addProduct(
              outletId: 'outlet-id',
              name: 'Kopi Susu',
              category: drink,
              price: 18000,
            );
        final updated = await container
            .read(posCatalogProvider.notifier)
            .updateProduct(
              product: created,
              outletId: 'outlet-id',
              name: 'Kopi Susu Besar',
              category: snack,
              price: 22000,
            );

        expect(repository.createdCategoryId, 'drink-id');
        expect(repository.updatedCategoryId, 'snack-id');
        expect(updated.category, 'Snack');
        expect(updated.categoryId, 'snack-id');
      },
    );

    test('should store edited category and remap products', () async {
      await container.read(posCatalogProvider.notifier).load();

      final updated = await container
          .read(posCatalogProvider.notifier)
          .updateCategory(id: 'snack-id', oldName: 'Snack', name: 'Camilan');

      final state = container.read(posCatalogProvider);
      expect(repository.updatedCategoryName, 'Camilan');
      expect(updated.name, 'Camilan');
      expect(state.categories, contains('Camilan'));
      expect(state.categories, isNot(contains('Snack')));
      expect(
        state.products
            .singleWhere((product) => product.id == 'snack-product')
            .category,
        'Camilan',
      );
    });
  });

  group('NojposSessionState RBAC', () {
    test(
      'should allow owner account to manage while active cashier is cashier',
      () {
        final state = NojposSessionState.initial().copyWith(
          account: const Employee(id: 'owner-id', name: 'Owner', role: 'owner'),
          cashier: const Employee(
            id: 'cashier-id',
            name: 'Kasir',
            role: 'cashier',
          ),
        );

        expect(state.canManageMasterData, true);
      },
    );

    test('should keep cashier account read-only', () {
      final state = NojposSessionState.initial().copyWith(
        account: const Employee(
          id: 'cashier-id',
          name: 'Kasir',
          role: 'cashier',
        ),
        cashier: const Employee(
          id: 'owner-id',
          name: 'Owner PIN',
          role: 'owner',
        ),
      );

      expect(state.canManageMasterData, false);
    });
  });
}

class _FakeProductRepository implements ProductRepository {
  int loadCalls = 0;
  String? createdCategoryId;
  String? updatedCategoryId;
  String? updatedCategoryName;

  @override
  Future<ProductCatalog> getCatalog({String? search, String? category}) async {
    loadCalls += 1;
    return const ProductCatalog(
      categoryItems: [
        ProductCategory(id: '__system_all', name: 'Semua'),
        ProductCategory(id: '__system_favorite', name: 'Favorit'),
        ProductCategory(id: 'drink-id', name: 'Minuman'),
        ProductCategory(id: 'snack-id', name: 'Snack'),
      ],
      products: [
        Product(
          id: 'snack-product',
          name: 'Keripik',
          category: 'Snack',
          categoryId: 'snack-id',
          price: 12000,
          imageUrl: '',
        ),
      ],
    );
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
    createdCategoryId = productCategoryId;
    return Product(
      id: 'created-product',
      name: name,
      category: category ?? 'Lainnya',
      categoryId: productCategoryId,
      price: price,
      imageUrl: '',
    );
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
    updatedCategoryId = productCategoryId;
    return Product(
      id: id,
      name: name,
      category: category ?? 'Lainnya',
      categoryId: productCategoryId,
      price: price,
      imageUrl: '',
    );
  }

  @override
  Future<ProductCategory> createCategory({required String name}) async {
    return ProductCategory(id: 'new-category-id', name: name);
  }

  @override
  Future<ProductCategory> updateCategory({
    required String id,
    required String name,
  }) async {
    updatedCategoryName = name;
    return ProductCategory(id: id, name: name);
  }
}
