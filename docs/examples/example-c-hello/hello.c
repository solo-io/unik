/*
#include <sys/time.h>
#include <stdlib.h>
#include <stdio.h>
#include <math.h>

int Fibonacci(int n)
{
   if ( n < 2 )
      return n;
   else
      return ( Fibonacci(n-1) + Fibonacci(n-2) );
} 

// Return 1 if the difference is negative, otherwise 0. 
int timeval_subtract(struct timeval *result, struct timeval *t2, struct timeval *t1)
{
    long int diff = (t2->tv_usec + 1000000 * t2->tv_sec) - (t1->tv_usec + 1000000 * t1->tv_sec);
    result->tv_sec = diff / 1000000;
    result->tv_usec = diff % 1000000;

    return (diff<0);
}

void timeval_print(struct timeval *tv)
{
    char buffer[30];
    time_t curtime;

    printf("%ld.%06ld", tv->tv_sec, tv->tv_usec);
    curtime = tv->tv_sec;
    strftime(buffer, 30, "%m-%d-%Y  %T", localtime(&curtime));
    printf(" = %s.%06ld\n", buffer, tv->tv_usec);
}

int main()
{
    struct timeval tvBegin, tvEnd, tvDiff;

    // begin
    gettimeofday(&tvBegin, NULL);
    timeval_print(&tvBegin);

    // lengthy operation
    int n = 46;
    printf("fibonacci number %d is: %d\n", n, Fibonacci(n) );

    //end
    gettimeofday(&tvEnd, NULL);
    timeval_print(&tvEnd);

    // diff
    timeval_subtract(&tvDiff, &tvEnd, &tvBegin);
    printf("%ld.%06ld\n", tvDiff.tv_sec, tvDiff.tv_usec);

    return 0;
}
*/

#include <stdio.h>

int main(){
	printf("hihi\n");
}
