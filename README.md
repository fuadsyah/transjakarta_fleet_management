# Fleet Management Backend

Sistem manajemen armada yang dapat menerima data lokasi kendaraan melalui MQTT, menyimpan ke PostgreSQL, menyediakan REST API, dan menggunakan RabbitMQ untuk event geofence.

## Teknologi yang Digunakan

- **Golang** - Backend service
- **MQTT (Eclipse Mosquitto)** - Menerima data lokasi kendaraan
- **PostgreSQL** - Database penyimpanan lokasi
- **RabbitMQ** - Event processing untuk geofence
- **Docker** - Container orchestration


## Struktur Project

```
.
├── cmd/
│   ├── publisher/     # MQTT mock data publisher
│   │   └── main.go
│   ├── server/        # Main backend server
│   │   └── main.go
│   └── worker/        # RabbitMQ consumer worker
│       └── main.go
├── internal/
│   ├── api/           # REST API router
│   ├── config/        # Configuration
│   ├── database/      # PostgreSQL connection
│   ├── geofence/      # Geofence checker
│   ├── handlers/      # HTTP handlers
│   ├── models/        # Data models
│   ├── mqtt/          # MQTT subscriber
│   ├── rabbitmq/      # RabbitMQ publisher & consumer
│   └── repository/    # Database repository
├── mosquitto/
│   └── config/        # Mosquitto configuration
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## Cara Menjalankan

### Prasyarat

- Docker dan Docker Compose terinstall
- Port yang tersedia: 1883 (MQTT), 5432 (PostgreSQL), 5672 & 15672 (RabbitMQ), 3000 (API)

### Menjalankan dengan Docker Compose

1. Clone repository:
```bash
git clone https://github.com/fuadsyah/transjakarta_fleet_management.git
cd transjakarta_fleet_management
```

2. Jalankan semua service:
```bash
docker-compose up --build
```

3. Untuk menjalankan di background:
```bash
docker-compose up --build -d
```

4. Melihat logs:
```bash
# Semua service
docker-compose logs -f

# Service tertentu
docker-compose logs -f server
docker-compose logs -f worker
docker-compose logs -f publisher
```

5. Menghentikan service:
```bash
docker-compose down
```

## API Endpoints

### Health Check
```
GET /health
```

Response:
```json
{
  "status": "healthy"
}
```

### Mendapatkan Lokasi Terakhir Kendaraan
```
GET /vehicles/{vehicle_id}/location
```

Response:
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

### Mendapatkan Riwayat Lokasi
```
GET /vehicles/{vehicle_id}/history?start={start_timestamp}&end={end_timestamp}
```

Response:
```json
[
  {
    "vehicle_id": "B1234XYZ",
    "latitude": -6.2088,
    "longitude": 106.8456,
    "timestamp": 1715000000
  },
  {
    "vehicle_id": "B1234XYZ",
    "latitude": -6.2089,
    "longitude": 106.8457,
    "timestamp": 1715000002
  }
]
```

## Konfigurasi

Konfigurasi dilakukan melalui environment variables. Lihat dalam file /internal/config/config.go

## Geofence Configuration

Default geofence dikonfigurasi di stasiun Bundaran HI
- Latitude: -6.1938148
- Longitude: 106.8230342
- Radius: 50 meter

## MQTT Topic

Data lokasi diterima melalui topic:
```
/fleet/vehicle/{vehicle_id}/location
```

Format pesan:
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

## RabbitMQ Configuration

- **Exchange**: fleet.events (type: direct)
- **Queue**: geofence_alerts
- **Routing Key**: geofence.entry

Format pesan geofence event:
```json
{
  "vehicle_id": "B1234XYZ",
  "event": "geofence_entry",
  "location": {
    "latitude": -6.2088,
    "longitude": 106.8456
  },
  "timestamp": 1715003456
}
```

## Testing

### Menggunakan curl

1. Health check:
```bash
curl http://localhost:3000/health
```

2. Get latest location:
```bash
curl http://localhost:3000/vehicles/B1234XYZ/location
```

3. Get location history:
```bash
curl "http://localhost:3000/vehicles/B1234XYZ/history?start=0&end=9999999999"
```

### Menggunakan Postman

Import file `postman_collection.json` ke Postman untuk testing API.
