docker run -v api_data:/data/db --name "mongodb" -e "MONGO_INITDB_ROOT_USERNAME=root" -e "MONGO_INITDB_ROOT_PASSWORD=example" -p "27017:27017" mongo