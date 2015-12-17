#include <stdlib.h>
#include <stdio.h>
#include <pthread.h>

void *count();
pthread_mutex_t mutex1 = PTHREAD_MUTEX_INITIALIZER;
int counter = 0;

int main() {
    int rc1, rc2;
    pthread_t thread1, thread2;

    /* 创建线程，每个线程独立执行函数 */
    if ((rc1 = pthread_create(&thread1, NULL, &count, NULL))) {
        printf("Thread create failed %d", rc1);
    }

    
    if ((rc2 = pthread_create(&thread1, NULL, &count, NULL))) {
        printf("Thread create failed %d", rc2);
    }

    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);

    return 0;
}

void *count() {
    pthread_mutex_lock(&mutex1);
    counter ++;
    printf("Counter value: %d \n", counter);
    pthread_mutex_unlock(&mutex1);
    return NULL;
}
