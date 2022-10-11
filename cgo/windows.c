/**
 * @file windows.c
 * @author {Do Huy Hoang} ({huyhoangdo0205@gmail.com})
 * @brief 
 * @version 1.0
 * @date 2022-10-11
 * 
 * @copyright Copyright (c) 2022
 * 
 */
#include <Windows.h>
#include <stdio.h>

void syncThreadDesktop(char** error) {
  HDESK hDesk = OpenInputDesktop(DF_ALLOWOTHERACCOUNTHOOK, FALSE, GENERIC_ALL);
  if(!hDesk) {
    DWORD err = GetLastError();

    char* message = (char*) malloc(100);
    memset(message,0,100);
    snprintf(message,100,"fail to open input desktop, error code %d",err);
    *error = message;
    return;
  }

  if(!SetThreadDesktop(hDesk)) {
    DWORD err = GetLastError();

    char* message = (char*) malloc(100);
    memset(message,0,100);
    snprintf(message,100,"fail to sync input desktop, error code %d",err);
    *error = message;
    return;
  }


  CloseDesktop(hDesk);
  *error = NULL;
  return;
}