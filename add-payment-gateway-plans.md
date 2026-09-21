act as senior software engineer again, saya ingin membuat sebuah adjustment lagi untuk POS ini. saya ingin membuat implementation integration with payment gateway. makesure code nya menggunakan architecture existing yaa.
1. plan saya ingin menggunakan midtrans untuk testing2 payment gateway integration ini, karena setau saya mereka cukup mudah untuk dibuat sebagai testing2 / sandbox. berikut adalah documentation link mereka:
    - backend: https://docs.midtrans.com/reference/backend-integration
    - frontend: https://docs.midtrans.com/reference/frontend-integration-overview
2. berikut adalah design/requirements detail yang saya inginkan: 
    - ubah logic di controller orders dan order_details, sesuaikan agar bisa menangani payment methods, bukan hanya Cash
    - ubah juga part di frontend nya karena akan call ke integration dengan midtrans
    - untuk sementara buat khusus untuk testing2 midtrans, buat environment variable untuk midtrans seperti api key, server key, dll
    - ketika sudah selesai, update documentation / swagger nya juga. 
    - oh iya, untuk backend, nanti akan ada endpoint /webhook dan saya ingin update status order otomatis berdasarkan response dari midtrans
    - related dengan implementation ini jadi sepertinya perlu membuat table baru untuk menyimpan transaction data yang berhubungan dengan midtrans, jadi seperti menyimpan historycal or log dari midtrans, jadi seperti menyimpan historycal / log dari midtrans yang berhubungan dengan order tersebut, jadi nanti ketika order sudah berhasil, status order berubah menjadi paid. bisa juga dibuat table semacam order_payments.
3. coba seperti itu dulu kira2 promptnya sudah cukup atau tidak.

    