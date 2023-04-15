## Interaction between main modules
```
|=====================|                           |=====================|   
|                     |                           |                     |
| Elevator Control    | <----------- requestCh -- | Request Control     |
|                     |                           |                     |
|                     | -- completedRequestCh --> |                     |
|                     |                           |                     |
|                     | -- sharedInfo ----------> |                     |
|=====================|                           |=====================|
```