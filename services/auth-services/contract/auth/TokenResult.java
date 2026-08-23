package auth;

import java.io.Serializable;

public class TokenResult implements Serializable {

    private static final long serialVersionUID = 1L;

    private final boolean valid;
    private final String message;
    private final String userId;
    private final String role;

    public TokenResult(
            boolean valid,
            String message,
            String userId,
            String role
    ) {
        this.valid = valid;
        this.message = message;
        this.userId = userId;
        this.role = role;
    }

    public boolean isValid() {
        return valid;
    }

    public String getMessage() {
        return message;
    }

    public String getUserId() {
        return userId;
    }

    public String getRole() {
        return role;
    }
}