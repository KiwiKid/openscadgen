include <BOSL2/std.scad>;
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

  renderType = "bottom_gap";


module wedge_with_box (height=100, wedge_height =60, width=100, depth = 2, sphere_radius =3, box_height = 14) {

difference(){
union(){
     difference(){
     union(){
     // wedge 
    hull(){

    // top left
    sphere(sphere_radius);
    
    // bottom left
    down(wedge_height)
    sphere(sphere_radius);
    
    back(height)
    sphere(sphere_radius);
    
    left(width)
    sphere(sphere_radius);
    
    back(depth)
    sphere(sphere_radius);
    
    left(width)
    back(depth)
    sphere(sphere_radius);
    
    // forward left
   left(width)
   back(depth)
   fwd(depth)
   sphere(sphere_radius);
    
   down(wedge_height)
   left(width)
  // back(height)
    sphere(sphere_radius);
    
   left(depth)
   back(depth) 
   down(depth)
     sphere(sphere_radius);
            
    left(width)
    back(height)
    sphere(sphere_radius);
    }

  // top box
   hull() {
        // middle back left
         sphere(sphere_radius);
         
           up(box_height)
            sphere(sphere_radius);
       
            left(width)
            up(box_height)
            sphere(sphere_radius);
            
            left(width)
            sphere(sphere_radius);
            
            
            back(height)
            up(box_height)
            sphere(sphere_radius);
            
            back(height)
            sphere(sphere_radius);
            
            
            left(width)
            back(height)
            up(box_height)
            sphere(sphere_radius);
            
            left(width)
            back(height)
            sphere(sphere_radius);
        }
    }
        
        
        
     cylider_radius_1 = 45;
      
     cylider_drop_1 = 2;
     
        cylider_radius_2 = 27;
 cylider_drop_2 = 16;
 
 

 cylider_radius_3 = 40;
   cylider_drop_3 = 8;

        left(width/2)
        back(height/2+1)
        down(cylider_drop_1)
        cylinder(20, cylider_radius_1,cylider_radius_1);
        
        
                left(width/2)
        back(height/2+1)
        down(cylider_drop_2)
        cylinder(20, cylider_radius_2,cylider_radius_2);
        
                        left(width/2)
        back(height/2+1)
        down(cylider_drop_3 )
        cylinder(20, cylider_radius_3,cylider_radius_3);
 
 }
        clip_off_set_z = 50;
        clip_width = 21;
        clip_height = 100;
        
        clip_cutout_size = [10000,14,170];
        clip_cutout_y_offset = 10;
        
        // back clip
        translate([0,0,clip_off_set_z])
        difference(){
         hull() {
        // middle back left
        fwd(clip_width)
        down(clip_height)
            sphere(sphere_radius);
            
        down(clip_height)
            sphere(sphere_radius);
            
            
           sphere(sphere_radius);

                        
        fwd(clip_width)
        down(clip_height)
        left(width)
            sphere(sphere_radius);
            
       fwd(clip_width)
        left(width)
            sphere(sphere_radius); 
            
            
        left(width)
            sphere(sphere_radius); 
            
        down(clip_height)
        left(width)
            sphere(sphere_radius);
            
        fwd(clip_width)
            sphere(sphere_radius);
         
         };
         
         down(clip_height)
         fwd(clip_cutout_y_offset)
         cuboid(clip_cutout_size, edges=[TOP+FRONT+LEFT, TOP+FRONT+RIGHT, 
         TOP+FRONT+LEFT, TOP+BACK+LEFT], rounding=3);
         
         }
         }
         
         
         phone_holder_rotate = [0,80,25];
         phone_holder_translate = [-32,60,65];

         phone_holder_size = [120,160,20];
         
         translate(phone_holder_translate)
        rotate(phone_holder_rotate)
        cuboid(phone_holder_size, rounding=2);
        
        
        mirror([-1,0,0]) translate([100,2,0]) {
            translate(phone_holder_translate)
            rotate(phone_holder_rotate)
            cuboid(phone_holder_size, rounding=2);
        }
        
        if(renderType == "bottom_gap"){
            move([-width/2,40,-70])
            cuboid([70,100,100], rounding=4);
            
             move([-width/2,height,-20])
            cuboid([40,180,80], rounding=4);
        }
        }
  }
  
  
  if(renderType == "horz-slice"){
      intersection(){
      wedge_with_box(); 
      fwd(500)
      left(500)
      cube([1000,1000,0.3]);
      }
  }else if(renderType =="vert-slice"){
       intersection(){
          wedge_with_box();
          rotate([90,0,90])
          fwd(500)
          left(500)
          down(50)
          cube([1000,1000,0.3]);
      }
  }else{ 
  
  wedge_with_box();
  }