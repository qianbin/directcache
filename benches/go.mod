module benches

go 1.12

require (
	github.com/VictoriaMetrics/fastcache v1.10.0
	github.com/allegro/bigcache/v3 v3.0.2
	github.com/coocood/freecache v1.2.1
	github.com/qianbin/directcache v0.9.1
)

replace github.com/qianbin/directcache => ../
