// API fitur purchasing: permukaan endpoint yang dipakai halaman
// Pembelian dan Supplier. Endpoint-nya sendiri tetap implementasi
// bersama di shared/api per domain (finance untuk supplier/PO/piutang,
// catalog untuk produk) — file ini hanya mengekspor ulang yang dipakai
// fitur purchasing, mengikuti pola fitur pos (satu pintu masuk API
// per fitur).
export {
  apiCancelPurchaseOrder, apiCreatePurchaseOrder, apiCreateSupplier, apiDeletePurchaseOrder, apiDeleteSupplier,
  apiGetPurchaseOrder, apiListPayables, apiListPurchaseOrders, apiListProducts, apiListSuppliers, apiOrderPurchaseOrder,
  apiReceivePurchaseOrder, apiUpdatePurchaseOrder, apiUpdateSupplier, fetchAll,
} from '../../shared/api'
export type {
  Payable, Product, PurchaseOrder, PurchaseOrderDetail, PurchaseOrderInput, PurchaseOrderReceiptInput, PurchaseOrderStatus, Supplier, SupplierInput,
} from '../../shared/api'
