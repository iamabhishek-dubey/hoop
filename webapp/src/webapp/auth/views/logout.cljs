(ns webapp.auth.views.logout
  (:require [webapp.components.headings :as h]))

(defn main []
  (set! (.. js/window -location -href)
        (str (. (. js/window -location) -origin) "/web-v1/login"))
  [:section {:class "antialiased min-h-screen bg-gray-100"}
   [:div {:class "px-x-large pb-x-large h-screen"}
    [:div {:class "p-regular"}
     [:figure {:class "w-36 px-small cursor-pointer"}
      [:img {:src "/web-v1/images/hoop-branding/PNG/hoop-symbol+text_black@4x.png"}]]]
    [:div {:class "w-full h-full bg-white rounded-lg"}
     [:div {:class "flex flex-col items-center"}
      [h/h2 "Logout successful!"
       {:class "mt-x-large px-large text-center"}]
      [:figure {:class "mt-x-large p-regular"}
       [:img {:src "/web-v1/images/illustrations/videogame.svg"
              :class "w-full"}]]]]]])
