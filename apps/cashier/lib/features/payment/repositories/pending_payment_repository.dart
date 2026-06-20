import '../../../core/network/api_client.dart';
import '../../transactions/repositories/transaction_repository.dart';

abstract interface class PendingPaymentRepository {
  Future<CheckoutTransaction> getTransaction(String transactionId);
}

class ApiPendingPaymentRepository implements PendingPaymentRepository {
  const ApiPendingPaymentRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<CheckoutTransaction> getTransaction(String transactionId) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/transactions/$transactionId',
    );
    return CheckoutTransaction.fromJson(response.data);
  }
}
