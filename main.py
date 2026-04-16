from archive import load_cities, load_officers
from ledger import connect, init_db, add_officer, get_all_officers, add_city, get_all_cities
from app import SovereignApp


def main() -> None:
    conn = connect()
    init_db(conn)

    if not get_all_officers(conn):
        for officer in load_officers():
            add_officer(conn, officer)
    if not get_all_cities(conn):
        for city in load_cities():
            add_city(conn, city)

    SovereignApp(conn).run()


if __name__ == "__main__":
    main()
