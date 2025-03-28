# backend.wemaps
Backend en Golang de Wemaps


Carga ETL

Crear un endpoint que gatille la tarea, esta tarea es unica, y si es gatillada por segunda vez se notificara su estado
Los estados de la tarea sera
1) Iniciando : Aca la tarea esta empezando a cargar, previo a la ejecucion de la consulta
Response :
{
    "Status":"initial",
    "detail":{
        "idtask":"HASH(timestamp + definicion desde requests)"
    }
}

2) En proceso : Aca la tarea se encuentra corriendo, es decir, aun esta desplegando data
Response :
{
    "Status":"in process",
    "detail":{
        "idtask":"HASH(timestamp + definicion desde requests)",
        "record_process":"10"
    }
}
3) Error: Aca la tarea es interrumpida por cualquier tipo de error
{
    "Status":"error",
    "detail":{
        "idtask":"HASH(timestamp + definicion desde requests)",
        "record_process":"10",
        "id_error":"500",
        "message":"Internal Error"
    }
}
4) Finalizado: La tarea a finalizado, entregando un resultado.
{
    "Status":"finish",
    "detail":{
        "idtask":"HASH(timestamp + definicion desde requests)",
        "record_process":"1010000"
    }
}

