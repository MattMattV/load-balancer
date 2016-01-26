# load-balancer (version HTTP)

[![Go Report Card](http://goreportcard.com/badge/MattMattV/load-balancer)](http://goreportcard.com/report/MattMattV/load-balancer)

Ce programme écrit en Go va récupérer l'état de tous les containers Docker sur la machine hôte *via* le container cAdvisor situé à l'adresse contenue dans la variable `MONITOR`.

Il n'y a qu'un seul point d'entrée : `/`

Le load-balancer va interroger cAdvisor et va donner au client le nom du serveur le moins chargé. Si une erreur se produit le code 500 sera renvoyé.

Le load-balancer va automatiquement mettre à jour sa liste de containers à équilibrer, grâce à une goroutine afin de délivrer un meilleur service.

3 variables d'environnement sont nécessaires au bon fonctionnement de l'application :
* `FILTER` : le "mot-clé" dont on va se servir pour filtrer les containers dont on veut équilibrer la charge 
* `MONITOR` : l'URL vers le container cAdvisor auprès duquel on va récupérer les données
* `HTTP_PORT` : le port sur lequel on va pouvoir contacter le load-balancer
