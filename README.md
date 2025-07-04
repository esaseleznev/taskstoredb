TaskStoreDB is a distributed key-value database optimized for task storage and management. It implements a cluster-aware architecture where tasks are distributed across nodes using consistent hashing through the hashring library. The system provides HTTP-based APIs for task operations and supports both local storage via LevelDB and distributed operations across cluster nodes.

Key capabilities include task lifecycle management, owner-based task assignment, group-based task organization, bulk operations, and search functionality across distributed nodes. The system implements CQRS (Command Query Responsibility Segregation) to separate write operations (app.Commands) from read operations (app.Queries).


[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/esaseleznev/taskstoredb)
