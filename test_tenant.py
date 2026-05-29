import requests
import hashlib
import json

# Log in
login_res = requests.post('http://localhost:8080/api/admin/login', json={'username': 'admin', 'password': 'admin2026'})
token = login_res.json()['token']

# Create tenant
headers = {'Authorization': f'Bearer {token}'}
create_res = requests.post('http://localhost:8080/api/admin/tenants', json={'name': 'TestTenantPY'}, headers=headers)
data = create_res.json()
print("Create response:", data)

if 'raw_api_key' in data:
    raw_key = data['raw_api_key']
    expected_hash = hashlib.sha256(raw_key.encode('utf-8')).hexdigest()
    print("Python calculated hash:", expected_hash)
