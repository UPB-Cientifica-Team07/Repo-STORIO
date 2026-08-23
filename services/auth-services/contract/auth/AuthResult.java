package auth;

import java.io.Serializable;

public class AuthResult implements Serializable {

    private static final long serialVersionUID = 1L;

    private final boolean success;
    private final String message;
    private final String userId;
    private final String username;
    private final String role;
    private final String token;

    public AuthResult(
            boolean success,
            String message,
            String userId,
            String username,
            String role,
            String token
    ) {
        this.success = success;
        this.message = message;
        this.userId = userId;
        this.username = username;
        this.role = role;
        this.token = token;
    }

    public boolean isSuccess() {
        return success;
    }

    public String getMessage() {
        return message;
    }

    public String getUserId() {
        return userId;
    }

    public String getUsername() {
        return username;
    }

    public String getRole() {
        return role;
    }

    public String getToken() {
        return token;
    }
}