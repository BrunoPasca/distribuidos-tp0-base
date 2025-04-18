# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```


## Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).


### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `


### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Ejercicio N°6:

#### Cliente:
Se modificó al cliente para que pueda enviar múltiples apuestas en un mismo mensaje (batch). El cliente ahora obtiene sus apuestas de un archivo CSV que se encuentra montado mediante un volumen Docker (cada cliente tiene su propio archivo correspondiente a su agencia, siguiendo el formato `.data/agency-{N}.csv`).

El cliente posee una configuración que determina la cantidad máxima de apuestas que pueden entrar en un batch, calculada para no exceder los 8kB por paquete. El cliente envía múltiples batches hasta completar todas las apuestas del archivo.

##### Formato de mensaje Cliente → Servidor (Batch de apuestas)
- **Header**:
  - 4 bytes: Longitud total del mensaje (incluyendo header y payload)
  - 1 byte: Tipo de mensaje (1 para batch de apuestas)
- **Payload**:
  - Datos de múltiples apuestas separadas por '\n' con el formato:
  - `clientId|nombre|apellido|documento|fechaNacimiento|numeroApostado`

#### Servidor:
El servidor procesa las apuestas recibidas en lotes, validando cada una de ellas. Si todas las apuestas son válidas, almacena la información y envía una confirmación al cliente. Si al menos una apuesta es inválida, ninguna se almacena y se informa al cliente del error.

##### Formato de respuesta Servidor → Cliente (Batch de apuestas)
- **Header**:
  - 2 bytes: Longitud de la respuesta
  - 1 byte: Tipo de mensaje (1 para respuesta a batch de apuestas)
- **Payload**:
  - `codigoRespuesta|cantidadApuestas`
  - donde codigoRespuesta es 0 para éxito o 1 para error

Cuando todas las apuestas del batch se procesan correctamente, el servidor registra: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de error con alguna apuesta, registra: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).

## Solucion

### Ejercicio 1:

Se generó un script bash `generar-compose.sh` que permite crear un archivo de Docker Compose con la cantidad de clientes especificada por línea de comandos. El script toma dos parámetros:

1. `nombre_archivo`: El nombre del archivo Docker Compose a generar
2. `cantidad_clientes`: La cantidad de clientes a incluir en la configuración

El script valida los parámetros de entrada y genera un archivo YAML que incluye:
- Un servicio para el servidor
- El número especificado de servicios cliente (client1, client2, etc.)
- Una configuración de red compartida para todos los servicios (propuesta por la catedra)

La solución incluye variables de entorno predeterminadas y volúmenes para la configuración, siguiendo la estructura del proyecto base.

Ejemplo de uso:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

Este comando generará una configuración de Docker Compose con un servidor y 5 clientes, guardándola en el archivo `docker-compose-dev.yaml`.

### Ejercicio 2:

Se implementó un sistema de inyección de configuración mediante volúmenes Docker para evitar la reconstrucción de imágenes cuando solo se modifican parámetros de configuración.

### Ejercicio 3:

Se desarrolló un script bash `validar-echo-server.sh` para verificar el funcionamiento del echo server sin instalar netcat en la máquina host ni exponer puertos del servidor. La solución:

1. Crea un contenedor temporal utilizando una imagen de Alpine ya existente, nos conectamos a la misma red que el echo server
2. Utiliza netcat dentro del contenedor para enviar un mensaje de prueba al servidor

### Ejercicio 4:

Se implementó un mecanismo de terminación graceful para el cliente y el servidor que responde adecuadamente a señales SIGTERM. La implementación garantiza que todos los recursos del sistema (sockets de red, archivos abiertos, hilos de ejecución) sean liberados correctamente antes de finalizar la ejecución del programa principal.

### Ejercicio 5:

Modificacion logica de negocio:

#### Cliente:
El cliente emula a una agencia de quiniela. Recibe como variables de entorno  nombre, apellido, DNI, nacimiento, numero apostado. Y estos campos son utilizados para luego enviar una apuesta al servidor.

#### Servidor: 
Emula ser la central de Loteria Nacional. Recibe apuestas de los clientes y las almacena.

#### Mensajes:
Para implementar la comunicación entre cliente y servidor se diseñó un protocolo que garantiza la correcta transmisión de las apuestas. El protocolo para el envío de una apuesta individual es el siguiente:

##### Formato de mensaje Cliente → Servidor (Apuesta individual)
- **Header**: 
  - 4 bytes: Longitud total del mensaje (incluyendo header y payload)
  - 1 byte: Tipo de mensaje (0 para apuesta individual)
- **Payload**: 
  - Datos de la apuesta en formato: `clientId|nombre|apellido|documento|fechaNacimiento|numeroApostado`

##### Formato de respuesta Servidor → Cliente
- **Header**:
  - 2 bytes: Longitud de la respuesta
  - 1 byte: Tipo de mensaje (0 para respuesta a apuesta individual)
- **Payload**:
  - Código de resultado y datos confirmados: `codigoRespuesta|documento|numeroApostado`
  - donde codigoRespuesta es 0 para éxito o 1 para error

### Ejercicio 6:

#### Cliente:
Se modificó al cliente para que pueda enviar múltiples apuestas en un mismo mensaje (batch). El cliente ahora obtiene sus apuestas de un archivo CSV que se encuentra montado mediante un volumen Docker (cada cliente tiene su propio archivo correspondiente a su agencia, siguiendo el formato `.data/agency-{N}.csv`).

El cliente posee una configuración que determina la cantidad máxima de apuestas que pueden entrar en un batch, calculada para no exceder los 8kB por paquete. El cliente envía múltiples batches hasta completar todas las apuestas del archivo.

#### Servidor:
El servidor procesa las apuestas recibidas en lotes, validando cada una de ellas. Si todas las apuestas son válidas, almacena la información y envía una confirmación al cliente. Si al menos una apuesta es inválida, ninguna se almacena y se informa al cliente del error.

##### Formato de mensaje Cliente → Servidor (Batch de apuestas)
- **Header**:
  - 4 bytes: Longitud total del mensaje (incluyendo header y payload)
  - 1 byte: Tipo de mensaje (1 para batch de apuestas)
- **Payload**:
  - Datos de múltiples apuestas separadas por '\n' con el formato:
  - `clientId|nombre|apellido|documento|fechaNacimiento|numeroApostado`

##### Formato de respuesta Servidor → Cliente (Batch de apuestas)
- **Header**:
  - 2 bytes: Longitud de la respuesta
  - 1 byte: Tipo de mensaje (1 para respuesta a batch de apuestas)
- **Payload**:
  - `codigoRespuesta|cantidadApuestas`
  - donde codigoRespuesta es 0 para éxito o 1 para error

### Ejercicio 7:

#### Cliente:
Los clientes envian una notificacion al servidor cuando finaliaron de enviar todas las apuestas. Luego periodicamente consultan al servidor para informarse si la loteria ya ha sido realizada. En caso de que si haya sido realizada, imprimen a los ganadores, sino esperan y vuelven a preguntar.

Tome la decision de que los clientes consulten periodicamente al servidor porque no tenemos una unica conexion con un cliente, cada cliente hasta este punto viene: creando una conexion, enviando mensaje, recibiendo respuesta y luego cerrando la conexion. Creo que hasta ahora este fue la mejor manera de encarar el problema, ya que al servidor ser incapaz de responder multiples clientes al mismo tiempo, crear una conexion que dure para toda la comunicacion hubiese realentizado considerablemente lo que tarda un cliente en poder completar su objetivo. Ya que si un cliente estaba conectado, no podemos conectarnos al servidor ni enviar apuestas y tendriamos que esperar a que se libere.

#### Servidor:
El servidor espera una notificacion de todas las agencias (cantidad determinada por variable de entorno en el Compose). Una vez que todas las agencias le informaron que estan esperando el sorteo, este se realiza. El servidor verifica cada apuesta que tiene almacenada y valida si es o no una apuesta ganadora. Luego responde a todas las agencias con los ganadores que corresponden a esa agencia (una agencia no debe enterarse de los ganadores de otra agencia).

##### Formato de mensaje Cliente → Servidor (Notificación de finalización)
- **Header**:
  - 4 bytes: Longitud total del mensaje (incluyendo header y payload)
  - 1 byte: Tipo de mensaje (2 para notificación de finalización)
- **Payload**:
  - Identificador de la agencia

##### Formato de respuesta Servidor → Cliente (Notificación de finalización)
- **Header**:
  - 2 bytes: Longitud total del mensaje (incluyendo header y payload)
  - 1 byte: Tipo de mensaje (2 para notificación de finalización)
- **Payload**:
  - 0

##### Formato de mensaje Cliente → Servidor (Consulta de ganadores)
- **Header**:
  - 4 bytes: Longitud total del mensaje (incluyendo header y payload)
  - 1 byte: Tipo de mensaje (3 para consulta de resultados)
- **Payload**:
  - Identificador de la agencia

##### Formato de respuesta Servidor → Cliente (Consulta de ganadores)
- **Header**:
  - 2 bytes: Longitud de la respuesta
  - 1 byte: Tipo de mensaje (3 para respuesta a consulta de resultados)
- **Payload**:
  - Si el sorteo ya fue realizado: `0|cantidad_ganadores`
  - Si el sorteo no fue realizado aún: `1` (código de error indicando que no está listo)

### Ejercicio 8:
Se modificó el servidor para que soporte múltiples clientes en paralelo utilizando la biblioteca `multiprocessing` de Python.

La implementación incluye:

1. **Procesamiento en paralelo**: Para cada nueva conexión de cliente, se crea un proceso independiente que maneja toda la comunicación con ese cliente específico.

2. **Recursos compartidos**: Se utiliza un `Manager` para gestionar estructuras de datos compartidas entre los procesos:
   - `agencies_waiting`: Diccionario que registra qué agencias están esperando el sorteo
   - `winners`: Diccionario con los ganadores organizados por agencia
   - `lottery_performed`: Flag que indica si ya se realizó el sorteo

3. **Mecanismos de sincronización**:
   - `agencies_lock`: Protege el acceso al diccionario de agencias en espera
   - `winners_lock`: Protege el acceso al diccionario de ganadores
   - `reader_lock` y `writer_lock`: Protegen las operaciones de lectura y escritura en archivos compartidos

4. **Gestión de procesos**: Se implementa un sistema de limpieza de procesos terminados para liberar recursos del sistema.