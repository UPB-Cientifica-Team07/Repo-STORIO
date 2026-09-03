package auth;

import java.util.Hashtable;

import javax.naming.AuthenticationException;
import javax.naming.Context;
import javax.naming.NamingEnumeration;
import javax.naming.NamingException;
import javax.naming.directory.Attributes;
import javax.naming.directory.DirContext;
import javax.naming.directory.InitialDirContext;
import javax.naming.directory.SearchControls;
import javax.naming.directory.SearchResult;
import javax.naming.ldap.Rdn;

public class LdapDirectoryClient {

    private final String ldapUrl;
    private final String baseDn;
    private final String peopleBase;
    private final String groupsBase;

    public LdapDirectoryClient() {

        this.ldapUrl =
                System.getenv()
                        .getOrDefault(
                                "LDAP_URL",
                                "ldap://127.0.0.1:389"
                        );

        this.baseDn =
                System.getenv()
                        .getOrDefault(
                                "LDAP_BASE_DN",
                                "dc=upb-cientifica,dc=local"
                        );

        this.peopleBase =
                "ou=people," + baseDn;

        this.groupsBase =
                "ou=groups," + baseDn;
    }

    public DirectoryUser authenticate(
            String username,
            String password
    ) throws NamingException {

        if (
                username == null ||
                username.isBlank() ||
                password == null ||
                password.isBlank()
        ) {

            return null;
        }

        String userRdn =
                new Rdn(
                        "uid",
                        username
                ).toString();

        String userDn =
                userRdn
                        + ","
                        + peopleBase;

        Hashtable<String, Object> environment =
                baseEnvironment();

        environment.put(
                Context.SECURITY_AUTHENTICATION,
                "simple"
        );

        environment.put(
                Context.SECURITY_PRINCIPAL,
                userDn
        );

        environment.put(
                Context.SECURITY_CREDENTIALS,
                password
        );

        DirContext context = null;

        try {

            /*
             * El constructor realiza el LDAP bind.
             * Si la contraseña es incorrecta,
             * OpenLDAP lanza AuthenticationException.
             */
            context =
                    new InitialDirContext(
                            environment
                    );

            Attributes attributes =
                    context.getAttributes(
                            userDn,
                            new String[] {
                                    "uid",
                                    "employeeNumber"
                            }
                    );

            String uid =
                    attributeValue(
                            attributes,
                            "uid"
                    );

            String directoryId =
                    attributeValue(
                            attributes,
                            "employeeNumber"
                    );

            if (
                    uid == null ||
                    uid.isBlank() ||
                    directoryId == null ||
                    directoryId.isBlank()
            ) {

                throw new NamingException(
                        "Usuario LDAP sin uid o employeeNumber"
                );
            }

            String role =
                    resolveRole(
                            context,
                            userDn
                    );

            if (
                    role == null ||
                    role.isBlank()
            ) {

                throw new NamingException(
                        "Usuario LDAP sin rol autorizado"
                );
            }

            return new DirectoryUser(
                    directoryId,
                    uid,
                    role
            );

        } catch (
                AuthenticationException error
        ) {

            return null;

        } finally {

            closeQuietly(
                    context
            );
        }
    }

    public DirectoryUser findByDirectoryId(
            String directoryId
    ) throws NamingException {

        if (
                directoryId == null ||
                directoryId.isBlank()
        ) {
            return null;
        }

        DirContext context = null;

        try {

            /*
             * Consulta anónima.
             *
             * La ACL predeterminada de este
             * laboratorio permite lectura.
             */
            context =
                    new InitialDirContext(
                            baseEnvironment()
                    );

            SearchControls controls =
                    new SearchControls();

            controls.setSearchScope(
                    SearchControls.SUBTREE_SCOPE
            );

            controls.setReturningAttributes(
                    new String[] {
                            "uid",
                            "employeeNumber"
                    }
            );

            String filter =
                    "(&(objectClass=inetOrgPerson)"
                            + "(employeeNumber="
                            + escapeFilter(
                                    directoryId
                            )
                            + "))";

            NamingEnumeration<SearchResult> results =
                    context.search(
                            peopleBase,
                            filter,
                            controls
                    );

            try {

                if (
                        !results.hasMore()
                ) {
                    return null;
                }

                SearchResult result =
                        results.next();

                Attributes attributes =
                        result.getAttributes();

                String uid =
                        attributeValue(
                                attributes,
                                "uid"
                        );

                String id =
                        attributeValue(
                                attributes,
                                "employeeNumber"
                        );

                String userDn =
                        "uid="
                                + Rdn.escapeValue(
                                        uid
                                )
                                + ","
                                + peopleBase;

                String role =
                        resolveRole(
                                context,
                                userDn
                        );

                if (
                        role == null
                ) {
                    return null;
                }

                return new DirectoryUser(
                        id,
                        uid,
                        role
                );

            } finally {

                results.close();
            }

        } finally {

            closeQuietly(
                    context
            );
        }
    }

    private String resolveRole(
            DirContext context,
            String userDn
    ) throws NamingException {

        SearchControls controls =
                new SearchControls();

        controls.setSearchScope(
                SearchControls.ONELEVEL_SCOPE
        );

        controls.setReturningAttributes(
                new String[] {
                        "cn"
                }
        );

        String filter =
                "(&(objectClass=groupOfNames)"
                        + "(member="
                        + escapeFilter(
                                userDn
                        )
                        + "))";

        NamingEnumeration<SearchResult> results =
                context.search(
                        groupsBase,
                        filter,
                        controls
                );

        boolean admin = false;
        boolean researcher = false;
        boolean user = false;

        try {

            while (
                    results.hasMore()
            ) {

                SearchResult result =
                        results.next();

                String role =
                        attributeValue(
                                result.getAttributes(),
                                "cn"
                        );

                if (
                        "ADMIN".equals(
                                role
                        )
                ) {

                    admin = true;

                } else if (
                        "INVESTIGADOR".equals(
                                role
                        )
                ) {

                    researcher = true;

                } else if (
                        "USUARIO".equals(
                                role
                        )
                ) {

                    user = true;
                }
            }

        } finally {

            results.close();
        }

        /*
         * Precedencia explícita de roles.
         */
        if (admin) {
            return "ADMIN";
        }

        if (researcher) {
            return "INVESTIGADOR";
        }

        if (user) {
            return "USUARIO";
        }

        return null;
    }

    private Hashtable<String, Object>
    baseEnvironment() {

        Hashtable<String, Object> environment =
                new Hashtable<>();

        environment.put(
                Context.INITIAL_CONTEXT_FACTORY,
                "com.sun.jndi.ldap.LdapCtxFactory"
        );

        environment.put(
                Context.PROVIDER_URL,
                ldapUrl
        );

        environment.put(
                "com.sun.jndi.ldap.connect.timeout",
                "3000"
        );

        environment.put(
                "com.sun.jndi.ldap.read.timeout",
                "5000"
        );

        return environment;
    }

    private static String attributeValue(
            Attributes attributes,
            String name
    ) throws NamingException {

        if (
                attributes == null ||
                attributes.get(
                        name
                ) == null
        ) {

            return null;
        }

        Object value =
                attributes
                        .get(
                                name
                        )
                        .get();

        return value == null
                ? null
                : value.toString();
    }

    private static String escapeFilter(
            String value
    ) {

        StringBuilder escaped =
                new StringBuilder();

        for (
                char character
                        : value.toCharArray()
        ) {

            switch (
                    character
            ) {

                case '\\' ->
                        escaped.append(
                                "\\5c"
                        );

                case '*' ->
                        escaped.append(
                                "\\2a"
                        );

                case '(' ->
                        escaped.append(
                                "\\28"
                        );

                case ')' ->
                        escaped.append(
                                "\\29"
                        );

                case '\0' ->
                        escaped.append(
                                "\\00"
                        );

                default ->
                        escaped.append(
                                character
                        );
            }
        }

        return escaped.toString();
    }

    private static void closeQuietly(
            DirContext context
    ) {

        if (
                context == null
        ) {
            return;
        }

        try {

            context.close();

        } catch (
                NamingException ignored
        ) {
        }
    }

    public record DirectoryUser(
            String id,
            String username,
            String role
    ) {
    }
}
