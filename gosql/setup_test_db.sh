if ! psql gosql_test -c "SELECT 1" > /dev/null 2>&1 ; then
	echo "creating db"
	psql template1 -c "CREATE DATABASE gosql_test";
fi
