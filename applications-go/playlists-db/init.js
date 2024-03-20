db = connect("localhost:27017", "root", "password");

db.playlists.insertMany([
    {
        "id": "1",
        "name": "CI/CD",
        "videos": [{ "id": "OFgziggbCOg" }, { "id": "myCcJJ_Fk10" }, { "id": "2WSJF7d8dUg" }]
    },
    {
        "id": "2",
        "name": "K8s in the Cloud",
        "videos": [{ "id": "QThadS3Soig" }, { "id": "eyvLwK5C2dw" }]
    },
    {
        "id": "3",
        "name": "Storage and MessageBrokers",
        "videos": [{ "id": "JmCn7k0PlV4" }, { "id": "_lpDfMkxccc" }]
    },
    {
        "id": "4",
        "name": "K8s Autoscaling",
        "videos": [{ "id": "jM36M39MA3I" }, { "id": "FfDI08sgrYY" }]
    }
]);
