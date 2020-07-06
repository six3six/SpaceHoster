# SPACEHOSTER API

Based on gRPC and protobuf

## Roadmap
Sorted by priority

### Release 1
- [x] Basic authentication
- [x] Automatic VM creation
- [ ] VM modification
- [ ] VM delete
- [ ] VNC Proxy
- [ ] LDAP authentication
- [ ] User info modification
- [ ] User management


### Release 2
- [ ] Docker/Kubernetes deployment
- [ ] Virtual network management
- [ ] Two factor authentication

## Environment variables
```dotenv
MONGO_INITDB_ROOT_USERNAME=
MONGO_INITDB_ROOT_PASSWORD=
MONGO_HOST=
MONGO_PORT=
PROXMOX_HOST=
PROXMOX_API_PORT=
PROXMOX_USER=
PROXMOX_PASSWORD=
```

### Dev
```dotenv
MONGO_INITDB_ROOT_USERNAME=root
MONGO_INITDB_ROOT_PASSWORD=example
MONGO_HOST=localhost
MONGO_PORT=27017
PROXMOX_HOST=CHANGE_HOST
PROXMOX_API_PORT=8006
PROXMOX_USER=CHANGE_USERNAME@CHANGE_SCOPE
PROXMOX_PASSWORD=CHANGE_PASSWORD
```

## Run
```shell script
cd protocol
make
cd ..
go get
go install
source .env
api
```

## Login server

### Login
Get connection token

### Register
Register new login

### Logout
Remove connection token


## VM server
### FreeResources
Get free resources and total allocated 

### Create
Create new VM (if enough resources)

### Start
Start VM (owner & manager)

### Stop
Stop VM (owner & manager)

### Restart
Restart VM (owner & manager)

### Delete
Delete VM (owner only)