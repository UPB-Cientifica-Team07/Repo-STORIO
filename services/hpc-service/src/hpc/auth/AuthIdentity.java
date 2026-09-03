package hpc.auth;

public record AuthIdentity(
    boolean valid,
    String userId,
    String role,
    String message
) {
}
