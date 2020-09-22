================================================================
``griprus`` -- Grip and Logrus Bidirectional Compatibility Layer
================================================================

gripslog is a shim layer between Grip and Slog, allowing libraries written
with one logger to exist within applications that use the other, and also use
the underlying fundamentals (Senders and Loggers) within either higher level
interfaces.

The translation between message.Composers and logrus.Entry (and between
level.Priority and logrus.Level) is not lossless, logrus stack data is not
converted to grip equivalents, and grip has a more fine grained level system;
however, for the most common cases the conversion is quite robust.
