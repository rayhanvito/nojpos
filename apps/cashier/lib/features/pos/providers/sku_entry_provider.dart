import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/product.dart';
import '../repositories/product_repository.dart';

final skuEntryProvider = NotifierProvider<SkuEntryNotifier, SkuEntryState>(
  SkuEntryNotifier.new,
);

class SkuEntryState {
  const SkuEntryState({
    this.query = '',
    this.results = const [],
    this.isLoading = false,
    this.errorMessage,
    this.hasSearched = false,
  });

  final String query;
  final List<Product> results;
  final bool isLoading;
  final String? errorMessage;
  final bool hasSearched;

  SkuEntryState copyWith({
    String? query,
    List<Product>? results,
    bool? isLoading,
    String? errorMessage,
    bool clearError = false,
    bool? hasSearched,
  }) {
    return SkuEntryState(
      query: query ?? this.query,
      results: results ?? this.results,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
      hasSearched: hasSearched ?? this.hasSearched,
    );
  }
}

class SkuEntryNotifier extends Notifier<SkuEntryState> {
  @override
  SkuEntryState build() => const SkuEntryState();

  Future<void> search(String rawQuery) async {
    final query = rawQuery.trim();
    if (query.isEmpty) {
      state = const SkuEntryState();
      return;
    }

    state = state.copyWith(
      query: query,
      isLoading: true,
      clearError: true,
      hasSearched: true,
    );

    try {
      final catalog = await ref
          .read(productRepositoryProvider)
          .getCatalog(search: query);
      state = state.copyWith(
        results: catalog.products,
        isLoading: false,
        query: query,
        hasSearched: true,
      );
    } catch (error) {
      state = state.copyWith(
        results: const [],
        isLoading: false,
        errorMessage: _messageFor(error),
        query: query,
        hasSearched: true,
      );
    }
  }

  void clear() {
    state = const SkuEntryState();
  }
}

String _messageFor(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
