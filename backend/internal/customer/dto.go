package customer

// CustomerInput adalah masukan tervalidasi untuk membuat/mengubah
// pelanggan (dipindah verbatim dari service.CustomerInput lama).
type CustomerInput struct {
	Name    string
	Phone   string
	Email   string
	Address string
}

// customerReq adalah bentuk JSON permintaan pelanggan (customerReq
// lama di paket handler); tag JSON tidak berubah.
type customerReq struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func (r customerReq) input() CustomerInput {
	return CustomerInput{Name: r.Name, Phone: r.Phone, Email: r.Email, Address: r.Address}
}
