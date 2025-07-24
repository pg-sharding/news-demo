#!/bin/bash

echo "Waiting for PostgreSQL shards..."
until pg_isready -h localhost -p 5550 -U user1 >/dev/null 2>&1; do 
    sleep 1
done
until pg_isready -h localhost -p 5551 -U user1 >/dev/null 2>&1; do 
    sleep 1
done

echo "Waiting for SPQR router..."
until pg_isready -h localhost -p 16432 -U user1 -d db1 >/dev/null 2>&1; do 
    sleep 1
done

echo "All services ready!"
