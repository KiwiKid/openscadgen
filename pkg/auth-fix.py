def name_callback(name, email):
    if not name or not email:
        return "Unknown", "unknown@example.com"
    if name == "greg-cumming-le":
        return "KiwiKid", "5018297+KiwiKid@users.noreply.github.com"
    return name, email
