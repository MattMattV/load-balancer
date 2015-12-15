# load-balancer (version HTTP)

Ce programme écrit en Go va récupérer l'état de tous les containers Docker sur la machine hôte *via* le container cAdvisor situé à l'adresse contenue dans la variable `monitor`.

Le client n'a qu'à interroger le load-balancer et tout le trafic TCP passera au travers du load-balancer en étant redirigé vers le serveur le moins chargé.
