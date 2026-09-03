#include <mpi.h>
#include <stdio.h>
#include <unistd.h>

int main(int argc, char **argv)
{
    int rank;
    int size;
    char hostname[256];

    MPI_Init(&argc, &argv);

    MPI_Comm_rank(
        MPI_COMM_WORLD,
        &rank
    );

    MPI_Comm_size(
        MPI_COMM_WORLD,
        &size
    );

    gethostname(
        hostname,
        sizeof(hostname)
    );

    printf(
        "Proceso %d de %d ejecutándose en %s\n",
        rank,
        size,
        hostname
    );

    MPI_Finalize();

    return 0;
}
