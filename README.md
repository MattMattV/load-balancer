# load-balancer (version HTTP)

Ce programme écrit en Go va récupérer l'état de tous les containers Docker sur la machine hôte *via* le container cAdvisor situé à l'adresse contenue dans la variable `monitor`.

On accède aux données *via* 2 points d'entrée :
* /view
* /redirect


`/view` va uniquement afficher le container qui a le plus de mémoire vive libre

`/redirect` va directement rediriger le client vers le serveur le moins chargé