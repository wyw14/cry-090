INSERT INTO venues(id,name,district,address,latitude,longitude,distance_bucket,approved)
VALUES ('venue-demo','Bailando Studio','徐汇','demo address',31.188,121.437,1,true)
ON CONFLICT (id) DO NOTHING;
