#include <stdlib.h>

#include <gst/gst.h>

#define MAX_ROUND 100

int
main (int argc, char **argv)
{
  GstElement *pipe, *filter,*filter2;
  GstCaps *caps;
  gint width, height, framerate = 1;
  gint xdir, ydir;
  gint round;
  GstMessage *message;

  gst_init (&argc, &argv);

  pipe = gst_parse_launch_full ("d3d11screencapturesrc ! video/x-raw(memory:D3D11Memory) ! capsfilter name=filter ! queue ! d3d11convert ! capsfilter name=filter2 ! queue !"
             "autovideosink", NULL, GST_PARSE_FLAG_NONE, NULL);
  g_assert (pipe != NULL);

  filter = gst_bin_get_by_name (GST_BIN (pipe), "filter");
  g_assert (filter);

  filter2 = gst_bin_get_by_name (GST_BIN (pipe), "filter2");
  g_assert (filter2);


  for (round = 0; round < MAX_ROUND; round++) {
    gchar *capsstr;
    /* we prefer our fixed width and height but allow other dimensions to pass
     * as well */

    capsstr = g_strdup_printf ("video/x-raw(memory:D3D11Memory),framerate=60/1");
    caps = gst_caps_from_string (capsstr);
    g_free (capsstr);
    g_object_set (filter, "caps", caps, NULL);
    gst_caps_unref (caps);

    if (round == 0)
      gst_element_set_state (pipe, GST_STATE_PLAYING);



    message =
        gst_bus_poll (GST_ELEMENT_BUS (pipe), GST_MESSAGE_ERROR,
        5000 * GST_MSECOND);
    if (message) {
      g_print ("got error %s at round %d \n",message->src->name,round);

      gst_message_unref (message);
    }
  }
  g_print ("done                    \n");

  gst_object_unref (filter);
  gst_element_set_state (pipe, GST_STATE_NULL);
  gst_object_unref (pipe);

  return 0;
}