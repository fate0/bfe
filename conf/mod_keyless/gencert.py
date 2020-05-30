#! /usr/bin/env python

import os

os.system("mkdir -p pub")
os.system("mkdir -p key")

os.system("cfssl gencert -initca json/ca_csr.json |cfssljson -bare ca")

print("generate keyless server client cert")

os.system('cfssl gencert -ca ca.pem -ca-key ca-key.pem -cn="www.keyless.com" -hostname="www.keyless.com" -config json/signing.json -profile client json/csr-ecdsa.json |cfssljson -bare client')
os.system('cfssl gencert -ca ca.pem -ca-key ca-key.pem -cn="www.keyless.com" -hostname="www.keyless.com" -config json/signing.json -profile server json/csr-ecdsa.json |cfssljson -bare server')

print("generate www certs")

for i in range(0, 10, 2):
    domain = f"www.{i}.com"
    os.system(f'cfssl gencert -ca ca.pem -ca-key ca-key.pem -cn="{domain}" -hostname="{domain}" -config json/signing.json -profile server json/csr-ecdsa.json |cfssljson -bare {domain}')
    os.system(f'mv {domain}-key.pem key/{domain}.key')
    os.system(f'mv {domain}.pem pub/{domain}.crt')

for i in range(1, 10, 2):
    domain = f"www.{i}.com"
    os.system(f'cfssl gencert -ca ca.pem -ca-key ca-key.pem -cn="{domain}" -hostname="{domain}" -config json/signing.json -profile server json/csr-rsa.json |cfssljson -bare {domain}')
    os.system(f'mv {domain}-key.pem key/{domain}.key')
    os.system(f'mv {domain}.pem pub/{domain}.crt')

os.system('rm *.csr')