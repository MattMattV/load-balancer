# load-balancer (version HTTP)

[![Go Report Card](http://goreportcard.com/report/Mistermatt007/load-balancer)](http://goreportcard.com/report/Mistermatt007/load-balancer)

Ce programme écrit en Go va récupérer l'état de tous les containers Docker sur la machine hôte *via* le container cAdvisor situé à l'adresse contenue dans la variable `monitor`.

Il n'y a qu'un seul point d'entrée : `/`

Le load-balancer va interroger cAdvisor et va donner au client le nom du serveur le moins chargé. Si une erreur se produit le code 500 sera renvoyé.
