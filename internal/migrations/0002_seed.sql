-- Data demo hasil porting dari FakeHomeDataSource dan FakeAuthRemoteDataSource Flutter.
INSERT INTO users (name, identifier, password_hash)
VALUES ('Kicau Mania', 'test@gmail.com', '$2a$10$GZo.EWIo5w6UVxx0YubDouDd48/Itn/qeDUg3nUlq27/Im2WWi8Hq')
ON CONFLICT (identifier) DO NOTHING;

INSERT INTO home_contents (id, payload)
VALUES (true, $json$
{
  "summary": {
    "location": "Jakarta Selatan",
    "liveCount": 14,
    "activeTicketCount": 2
  },
  "featured": {
    "title": "Piala Raja Nusantara 2026",
    "schedule": "Minggu, 6 Des • Batavia Arena, Jaksel",
    "distance": "3.2 KM",
    "prize": "Rp 50 Jt",
    "slots": "12 Gantangan",
    "attendees": {
      "initials": ["MK", "BN"],
      "others": 48
    }
  },
  "upcoming": {
    "type": "Latber Presisi",
    "title": "Latber Presisi Kacer",
    "time": "Besok, 14:00 WIB",
    "price": "Rp 75.000",
    "remainingTickets": 3
  },
  "stats": {
    "badge": "100% Resmi",
    "headline": "520+ Gantangan",
    "note": "Terverifikasi panitia & EO resmi.",
    "securityNote": "98% Transaksi Tiket Instan & Aman"
  },
  "activeTicket": {
    "event": "Grand Prix Murai Batu VIP #18",
    "schedule": "Sabtu, 12 Des • Masterpiece Arena",
    "number": "#G-18",
    "session": "Sesi A - 16G"
  },
  "schedules": [
    {
      "title": "Latpres Exclusive Cucak Ijo",
      "venue": "Arena Master Cilandak, Jaksel",
      "date": "Selasa, 24 Nov",
      "price": "Rp 100rb",
      "status": "available",
      "category": "cucakHijau"
    },
    {
      "title": "Festival Kicau Batavia",
      "venue": "Gantangan Taman Mini, Jaktim",
      "date": "Kamis, 26 Nov",
      "price": "Rp 150rb",
      "status": "almostGone",
      "category": "muraiBatu"
    },
    {
      "title": "Latber Kacer Mingguan",
      "venue": "Lapangan Kicau Pesanggrahan, Jaksel",
      "date": "Sabtu, 28 Nov",
      "price": "Rp 60rb",
      "status": "soldOut",
      "category": "kacer"
    },
    {
      "title": "Kenari Cup Kemang",
      "venue": "Kemang Bird Park, Jaksel",
      "date": "Minggu, 29 Nov",
      "price": "Rp 85rb",
      "status": "waiting",
      "category": "kenari"
    }
  ]
}
$json$)
ON CONFLICT (id) DO UPDATE SET payload = EXCLUDED.payload;
