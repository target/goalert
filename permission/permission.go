/*

Package permission handles checking and granting of permissions using context.Context.

A context can be granted User, System, or Service privileges using UserContext, SystemContext, or ServiceContext, respectively.

Data can be extracted using the appropriate method (e.g. UserID, ServiceID, etc...)

Context can then be validated using Checkers (e.g. like the User function) or by using LimitCheckAny and a number of Checkers together.

*/
package permission
