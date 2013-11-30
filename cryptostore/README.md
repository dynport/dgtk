# Crypter

## Goal

Store encrypted BLOBs of data for multiple users on a server. New users can be added by all existing users, BLOBs can be changed by all users.

## Create user

* All user data is stored in a user specific directory `$ROOT/users/<login>`
* Creating of users requires the login name and a user specific password
* A new 4096 bit RSA keypair is created, the public key is stored unencrypted, the privat key is encrypted with AES and the provided password

## Store BLOB for a specific user

* a new 32 byte secret AES key is created
* the BLOB is encrypted and stored with the generated key `$ROOT/users/<login>/data.<version>
* the generated key is encrypted with the public key of the user

## Read BLOB by user

* the private RSA key of the user is decrypted by the user provided password
* the secret key of the BLOB is decrypted with private RSA key
* the BLOB es decrypted withg the secret key


## Approach

All users have secret 32 byte keys which are provided with each request.

## Requirements

All stored BLOBs need to have some version (or checksum) in their names.
