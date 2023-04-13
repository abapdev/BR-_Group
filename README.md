# BR_Group
Тестовое задание для BR _Group на позицию Golang developer, разработчик backend (middle)

Описание

Требуется реализовать подключение к WebSocket биржи ascendex (https://ascendex.github.io/ascendex-pro-api/#websocket). Интерфейс подключения содержится в файле apiclient.go

Connection - функция должна реализовать подключение к WebSocket биржи. Если происходит проблемы с подключением, то должна вернуть ошибку

Disconnect - функция должна реализовывать отключение от WebSocket биржи.

SubscribeToChannel - функция должна реализовывать прослушивание канала WebSocket по получению BBO. Если происходит проблемы с прослушиванием, то должна вернуть ошибку и корректно отключиться от WebSocket.

ReadMessagesFromChannel - функция должна реализовывать чтение канала WebSocket о BBO, правильно преобразовывать данные и записывать их с chan.

WriteMessagesToChannel - функция должна реализовывать записи в канал WebSocket, чтобы поддерживать подключение открытым.

Требования

• Код выложен на github
• При решении использовать github.com/gorilla/websocket для работы с WebSocket
• Написать тесты
