## Interaction between main modules
```
|=====================|                           |=====================|   
|                     |                           |                     |
| Elevator Control    | <---------- requestsCh -- | Request Control     |
|                     |                           |                     |
|                     | -- completedRequestCh --> |                     |
|                     |                           |                     |
|                     | -- sharedInfo ----------> |                     |
|=====================|                           |=====================|
```
